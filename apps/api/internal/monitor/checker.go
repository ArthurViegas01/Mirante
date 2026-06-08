package monitor

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lumni/mirante/internal/platform/httpx"
)

// Checker runs a single probe against a service and returns a raw Sample.
type Checker interface {
	Check(ctx context.Context, svc *Service) Sample
}

type checker struct {
	fetcher *httpx.Fetcher
	dialer  *net.Dialer
}

// NewChecker builds the default checker with the monitor fetch policy (the
// owner's own infra, so private IPs are allowed).
func NewChecker() Checker {
	return &checker{
		fetcher: httpx.NewFetcher(httpx.Policy{AllowPrivateIPs: true, MaxBodyBytes: 32 << 10}),
		dialer:  &net.Dialer{},
	}
}

// Check applies the service's hard timeout and dispatches by kind.
func (c *checker) Check(ctx context.Context, svc *Service) Sample {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(svc.TimeoutMs)*time.Millisecond)
	defer cancel()
	if svc.Kind == KindHTTP {
		return c.checkHTTP(cctx, svc)
	}
	// tcp and db_ping are connectivity-only: a TCP connect, never a query.
	return c.checkTCP(cctx, svc)
}

func (c *checker) checkHTTP(ctx context.Context, svc *Service) Sample {
	res, err := c.fetcher.Get(ctx, svc.Target)
	if err != nil {
		return Sample{Responded: false, OK: false}
	}
	return Sample{
		Responded:  true,
		OK:         statusMatches(res.StatusCode, svc.ExpectedStatus),
		LatencyMs:  res.LatencyMs,
		StatusCode: res.StatusCode,
	}
}

func (c *checker) checkTCP(ctx context.Context, svc *Service) Sample {
	start := time.Now()
	conn, err := c.dialer.DialContext(ctx, "tcp", svc.Target)
	if err != nil {
		return Sample{Responded: false, OK: false}
	}
	_ = conn.Close()
	return Sample{Responded: true, OK: true, LatencyMs: int(time.Since(start).Milliseconds())}
}

// statusMatches checks a code against an expectation: a class like "2xx" or a
// comma list like "200,204".
func statusMatches(code int, expected string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" || expected == "2xx" {
		return code >= 200 && code < 300
	}
	if len(expected) == 3 && strings.HasSuffix(expected, "xx") {
		if class, err := strconv.Atoi(expected[:1]); err == nil {
			return code >= class*100 && code < (class+1)*100
		}
	}
	for _, part := range strings.Split(expected, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n == code {
			return true
		}
	}
	return false
}
