package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeResolver returns scripted answers so DNS-rebind behavior is deterministic.
// answers holds successive responses (the last repeats); addrs is the simple
// single-answer form.
type fakeResolver struct {
	mu      sync.Mutex
	calls   int
	answers [][]net.IPAddr
	addrs   []net.IPAddr
	err     error
}

func (r *fakeResolver) LookupIPAddr(_ context.Context, _ string) ([]net.IPAddr, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if len(r.answers) > 0 {
		i := r.calls - 1
		if i >= len(r.answers) {
			i = len(r.answers) - 1
		}
		return r.answers[i], nil
	}
	return r.addrs, nil
}

func (r *fakeResolver) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func ipAddrs(ss ...string) []net.IPAddr {
	out := make([]net.IPAddr, len(ss))
	for i, s := range ss {
		out[i] = net.IPAddr{IP: net.ParseIP(s)}
	}
	return out
}

// A host resolving to loopback is refused at the dial — and only one resolution
// happens, so there is no check-then-dial window for a rebind.
func TestFetchRefusesPrivateIP(t *testing.T) {
	fr := &fakeResolver{addrs: ipAddrs("127.0.0.1")}
	f := NewFetcher(Policy{AllowPrivateIPs: false})
	f.resolver = fr

	_, _, err := f.Fetch(context.Background(), "http://evil.example/path")
	if !errors.Is(err, ErrBlockedIP) {
		t.Fatalf("want ErrBlockedIP, got %v", err)
	}
	if n := fr.count(); n != 1 {
		t.Fatalf("want exactly 1 DNS resolution (no check-then-dial window), got %d", n)
	}
}

// The cloud-metadata link-local address is blocked.
func TestFetchRefusesCloudMetadata(t *testing.T) {
	fr := &fakeResolver{addrs: ipAddrs("169.254.169.254")}
	f := NewFetcher(Policy{AllowPrivateIPs: false})
	f.resolver = fr

	_, _, err := f.Fetch(context.Background(), "https://metadata.example/")
	if !errors.Is(err, ErrBlockedIP) {
		t.Fatalf("want ErrBlockedIP, got %v", err)
	}
}

// A hostname that resolves to both a public decoy and a private target is
// rejected outright (never dials the public alternative).
func TestFetchRefusesMixedPublicPrivate(t *testing.T) {
	fr := &fakeResolver{addrs: ipAddrs("93.184.216.34", "10.0.0.5")}
	f := NewFetcher(Policy{AllowPrivateIPs: false})
	f.resolver = fr

	_, _, err := f.Fetch(context.Background(), "http://mixed.example/")
	if !errors.Is(err, ErrBlockedIP) {
		t.Fatalf("want ErrBlockedIP, got %v", err)
	}
}

// DNS rebind: the resolver answers a (non-routable) public IP first and loopback
// second. The old check-then-dial code would re-resolve and connect to loopback;
// the pinned dialer resolves exactly once, so the malicious second answer is
// never consulted.
func TestFetchPinsSingleResolutionNoRebind(t *testing.T) {
	fr := &fakeResolver{answers: [][]net.IPAddr{
		ipAddrs("203.0.113.10"), // TEST-NET-3: public, non-routable
		ipAddrs("127.0.0.1"),    // the rebind target, must never be used
	}}
	f := NewFetcher(Policy{AllowPrivateIPs: false})
	f.resolver = fr

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	_, _, err := f.Fetch(ctx, "http://rebind.example/")
	if err == nil {
		t.Fatal("expected a dial error to the pinned public IP, got nil")
	}
	if errors.Is(err, ErrBlockedIP) {
		t.Fatalf("the pinned IP was public and should not be policy-blocked: %v", err)
	}
	if n := fr.count(); n != 1 {
		t.Fatalf("the second (malicious) DNS answer must never be consulted: want 1 lookup, got %d", n)
	}
}

// The fetcher dials exactly the resolved IP and reports it as ResolvedIP.
func TestFetchPinsAndReportsResolvedIP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	host, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("split server url: %v", err)
	}
	fr := &fakeResolver{addrs: ipAddrs(host)}
	f := NewFetcher(Policy{AllowPrivateIPs: true}) // monitor-style policy: loopback allowed
	f.resolver = fr

	res, body, err := f.Fetch(context.Background(), "http://service.local:"+port+"/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("unexpected response: status=%d body=%q", res.StatusCode, body)
	}
	if res.ResolvedIP != host {
		t.Fatalf("ResolvedIP: want %s, got %s", host, res.ResolvedIP)
	}
}

func TestFetchRejectsBadScheme(t *testing.T) {
	f := NewFetcher(Policy{})
	if _, _, err := f.Fetch(context.Background(), "ftp://example.com/x"); !errors.Is(err, ErrBadScheme) {
		t.Fatalf("want ErrBadScheme, got %v", err)
	}
}

func TestPickAllowedIP(t *testing.T) {
	cases := []struct {
		name         string
		ips          []net.IPAddr
		allowPrivate bool
		wantErr      bool
		want         string
	}{
		{"public allowed", ipAddrs("93.184.216.34"), false, false, "93.184.216.34"},
		{"loopback blocked", ipAddrs("127.0.0.1"), false, true, ""},
		{"private blocked", ipAddrs("10.1.2.3"), false, true, ""},
		{"link-local blocked", ipAddrs("169.254.169.254"), false, true, ""},
		{"private allowed when policy permits", ipAddrs("10.1.2.3"), true, false, "10.1.2.3"},
		{"any-private in set is blocked", ipAddrs("93.184.216.34", "192.168.0.1"), false, true, ""},
		{"cgnat blocked", ipAddrs("100.64.0.1"), false, true, ""},
		{"nat64 blocked", ipAddrs("64:ff9b::7f00:1"), false, true, ""},
		{"empty set errors", nil, false, true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ip, err := pickAllowedIP(c.ips, c.allowPrivate)
			if c.wantErr {
				if err == nil {
					t.Fatalf("want error, got ip=%v", ip)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ip.String() != c.want {
				t.Fatalf("want %s, got %s", c.want, ip)
			}
		})
	}
}
