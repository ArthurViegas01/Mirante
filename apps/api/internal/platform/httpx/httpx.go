// Package httpx is a policy-parameterized safe HTTP fetcher (ADR-0003). The
// monitor probes the owner's own infra (private IPs allowed); job-link fetching
// (F3) will use the opposite policy (private IPs blocked + IP pinning against
// DNS rebind). F1 exercises the monitor policy.
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Policy controls a fetcher's trust posture.
type Policy struct {
	AllowPrivateIPs bool
	MaxBodyBytes    int64
}

// Errors.
var (
	ErrBadScheme = errors.New("unsupported URL scheme")
	ErrBlockedIP = errors.New("target IP is blocked by policy")
)

// Fetcher performs guarded GET requests.
type Fetcher struct {
	policy Policy
	client *http.Client
}

// NewFetcher builds a fetcher that never follows redirects.
func NewFetcher(p Policy) *Fetcher {
	if p.MaxBodyBytes <= 0 {
		p.MaxBodyBytes = 64 << 10
	}
	return &Fetcher{
		policy: p,
		client: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// Result is a successful probe outcome.
type Result struct {
	StatusCode int
	LatencyMs  int
	ResolvedIP string
}

// Get performs one GET. The context carries the timeout. The response body is
// drained up to MaxBodyBytes and discarded.
func (f *Fetcher) Get(ctx context.Context, rawURL string) (*Result, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrBadScheme
	}

	ip, err := f.resolveAndCheck(ctx, u.Hostname())
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mirante-Monitor/1.0")

	start := time.Now()
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, f.policy.MaxBodyBytes))

	return &Result{
		StatusCode: resp.StatusCode,
		LatencyMs:  int(time.Since(start).Milliseconds()),
		ResolvedIP: ip,
	}, nil
}

func (f *Fetcher) resolveAndCheck(ctx context.Context, host string) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !f.policy.AllowPrivateIPs && isPrivate(ip) {
			return "", ErrBlockedIP
		}
		return ip.String(), nil
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("resolve %s: no addresses", host)
	}
	for _, ip := range ips {
		if !f.policy.AllowPrivateIPs && isPrivate(ip) {
			return "", ErrBlockedIP
		}
	}
	return ips[0].String(), nil
}

func isPrivate(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
