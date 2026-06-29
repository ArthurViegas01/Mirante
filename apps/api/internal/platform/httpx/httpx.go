// Package httpx is a policy-parameterized safe HTTP fetcher (ADR-0003). The
// monitor probes the owner's own infra (private IPs allowed); job-link and
// GitHub fetching use the opposite policy (private IPs blocked + IP pinning
// against DNS rebind). The validated IP is the exact one dialed: resolution
// happens once, inside DialContext, so there is no check-then-dial window for a
// rebind (F2).
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

// Policy controls a fetcher's trust posture.
type Policy struct {
	AllowPrivateIPs bool
	MaxBodyBytes    int64
	UserAgent       string
}

// Errors.
var (
	ErrBadScheme = errors.New("unsupported URL scheme")
	ErrBlockedIP = errors.New("target IP is blocked by policy")
)

// resolver is the slice of net.Resolver this package needs; it is injectable so
// the DNS-rebind behavior can be tested deterministically. *net.Resolver (and
// net.DefaultResolver) satisfy it.
type resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// Fetcher performs guarded GET requests.
type Fetcher struct {
	policy   Policy
	client   *http.Client
	resolver resolver
}

// NewFetcher builds a fetcher that never follows redirects and dials only IPs
// allowed by the policy. The custom Transport.DialContext resolves the host and
// validates the chosen IP, then dials that exact IP — the address used on the
// wire is the one that passed the check, closing the DNS-rebind window.
func NewFetcher(p Policy) *Fetcher {
	if p.MaxBodyBytes <= 0 {
		p.MaxBodyBytes = 64 << 10
	}
	if p.UserAgent == "" {
		p.UserAgent = "Mirante-Monitor/1.0"
	}
	f := &Fetcher{policy: p, resolver: net.DefaultResolver}
	transport := &http.Transport{
		DialContext:           f.dialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          8,
	}
	f.client = &http.Client{
		Transport:     transport,
		Timeout:       30 * time.Second, // overall cap: dial + TLS + headers + body
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return f
}

// Result is a successful probe outcome.
type Result struct {
	StatusCode int
	LatencyMs  int
	ResolvedIP string
}

// Get performs one GET, draining and discarding the body (used by the monitor,
// which only cares about reachability/latency).
func (f *Fetcher) Get(ctx context.Context, rawURL string) (*Result, error) {
	res, _, err := f.do(ctx, rawURL, false)
	return res, err
}

// Fetch performs one GET and returns the response body (up to MaxBodyBytes). Used
// by job-link import to read a posting's HTML.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, []byte, error) {
	return f.do(ctx, rawURL, true)
}

func (f *Fetcher) do(ctx context.Context, rawURL string, keepBody bool) (*Result, []byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil, ErrBadScheme
	}

	// Capture the IP actually dialed (the pinned one) for the result; GotConn
	// fires for the connection the request used, even on reuse.
	var resolvedIP string
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			if info.Conn != nil {
				if host, _, e := net.SplitHostPort(info.Conn.RemoteAddr().String()); e == nil {
					resolvedIP = host
				}
			}
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", f.policy.UserAgent)

	start := time.Now()
	resp, err := f.client.Do(req)
	if err != nil {
		// A blocked IP surfaces as the dial error; unwrap the *url.Error so callers
		// can still match errors.Is(err, ErrBlockedIP).
		return nil, nil, unwrapURLError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	res := &Result{
		StatusCode: resp.StatusCode,
		LatencyMs:  int(time.Since(start).Milliseconds()),
		ResolvedIP: resolvedIP,
	}

	limited := io.LimitReader(resp.Body, f.policy.MaxBodyBytes)
	if keepBody {
		body, err := io.ReadAll(limited)
		if err != nil {
			return nil, nil, err
		}
		return res, body, nil
	}
	_, _ = io.Copy(io.Discard, limited)
	return res, nil, nil
}

// dialContext resolves the host, picks an allowed IP, and dials that exact IP —
// so the address validated is the address connected to (anti DNS-rebind).
func (f *Fetcher) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip, err := f.pinnedIP(ctx, host)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{Timeout: 10 * time.Second}
	return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
}

// pinnedIP returns the single IP to dial for host, after the policy check. A
// literal IP is checked directly; a hostname is resolved once here. If the policy
// forbids private IPs and the host resolves to any of them, the request is
// refused (ErrBlockedIP) rather than silently dialing a public alternative.
func (f *Fetcher) pinnedIP(ctx context.Context, host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !f.policy.AllowPrivateIPs && isPrivate(ip) {
			return nil, ErrBlockedIP
		}
		return ip, nil
	}
	ips, err := f.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", host, err)
	}
	return pickAllowedIP(ips, f.policy.AllowPrivateIPs)
}

// pickAllowedIP returns the first resolved address, but refuses the whole set if
// the policy blocks private IPs and any resolved address is private (a hostname
// that maps to both a public decoy and a private target is rejected outright).
func pickAllowedIP(ips []net.IPAddr, allowPrivate bool) (net.IP, error) {
	if len(ips) == 0 {
		return nil, errors.New("no addresses resolved")
	}
	if !allowPrivate {
		for _, ipa := range ips {
			if isPrivate(ipa.IP) {
				return nil, ErrBlockedIP
			}
		}
	}
	return ips[0].IP, nil
}

// Extra ranges net.IP.IsPrivate misses but that can route to internal/edge infra:
// CGNAT (RFC 6598) and the well-known NAT64 prefix (RFC 6052).
var (
	cgnatRange = mustCIDR("100.64.0.0/10")
	nat64Range = mustCIDR("64:ff9b::/96")
)

func mustCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}

func isPrivate(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() ||
		cgnatRange.Contains(ip) || nat64Range.Contains(ip)
}

// unwrapURLError peels a *url.Error so a policy rejection (ErrBlockedIP) raised
// inside DialContext stays matchable with errors.Is at the call site.
func unwrapURLError(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) && errors.Is(ue.Err, ErrBlockedIP) {
		return ue.Err
	}
	return err
}
