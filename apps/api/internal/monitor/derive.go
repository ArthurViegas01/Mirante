package monitor

import "fmt"

// Sample is the raw result of a single probe.
type Sample struct {
	Responded  bool // got a connection / HTTP response at all
	OK         bool // responded AND correct (status matched / connected cleanly)
	LatencyMs  int
	StatusCode int // HTTP status (0 for tcp/db_ping)
}

// Thresholds parameterize the state machine.
type Thresholds struct {
	DegradedMs int // a successful response slower than this is "degraded"
	N          int // consecutive failures required to declare "down" (anti-flap)
	K          int // consecutive successes required to recover from "down"
}

// DeriveInput is the previous state plus the new sample.
type DeriveInput struct {
	Prev            Status
	ConsecFailures  int
	ConsecSuccesses int
	Sample          Sample
	T               Thresholds
}

// DeriveResult is the new state, updated counters and whether a transition
// occurred (a transition is what raises an alert and emits a live event).
type DeriveResult struct {
	State           Status
	ConsecFailures  int
	ConsecSuccesses int
	Changed         bool
	Reason          string
	Outcome         Status // raw classification of THIS probe (up|degraded|down)
}

// Derive is the pure monitor state machine. It is deterministic and side-effect
// free so it can be exhaustively table-tested.
//
// Rules:
//   - A probe is a FAILURE if it did not respond or responded incorrectly
//     (wrong code / timeout). Otherwise it is a SUCCESS, classified as
//     "degraded" when slower than DegradedMs, else "up".
//   - "down" requires N consecutive failures (anti-flap). Until then the prior
//     state is held.
//   - Recovering from "down" requires K consecutive successes (hysteresis).
//   - The up<->degraded edge is immediate (both are healthy responses).
func Derive(in DeriveInput) DeriveResult {
	t := in.T
	if t.N < 1 {
		t.N = 1
	}
	if t.K < 1 {
		t.K = 1
	}

	failure := !in.Sample.Responded || !in.Sample.OK
	res := DeriveResult{
		ConsecFailures:  in.ConsecFailures,
		ConsecSuccesses: in.ConsecSuccesses,
	}

	switch {
	case failure:
		res.Outcome = StatusDown
	case in.Sample.LatencyMs > t.DegradedMs:
		res.Outcome = StatusDegraded
	default:
		res.Outcome = StatusUp
	}

	if failure {
		res.ConsecFailures = in.ConsecFailures + 1
		res.ConsecSuccesses = 0
		if res.ConsecFailures >= t.N {
			res.State = StatusDown
			res.Reason = fmt.Sprintf("%d consecutive failures", res.ConsecFailures)
		} else {
			res.State = holdState(in.Prev)
			res.Reason = fmt.Sprintf("failure %d/%d", res.ConsecFailures, t.N)
		}
	} else {
		res.ConsecSuccesses = in.ConsecSuccesses + 1
		res.ConsecFailures = 0
		target := res.Outcome // up or degraded
		if in.Prev == StatusDown {
			if res.ConsecSuccesses >= t.K {
				res.State = target
				res.Reason = fmt.Sprintf("recovered after %d successes", res.ConsecSuccesses)
			} else {
				res.State = StatusDown
				res.Reason = fmt.Sprintf("recovering %d/%d", res.ConsecSuccesses, t.K)
			}
		} else {
			res.State = target
			if target == StatusDegraded {
				res.Reason = fmt.Sprintf("slow response %dms", in.Sample.LatencyMs)
			} else {
				res.Reason = "healthy"
			}
		}
	}

	res.Changed = res.State != in.Prev
	return res
}

// holdState is the state to keep while failures have not yet crossed N.
func holdState(prev Status) Status {
	if prev == "" {
		return StatusUnknown
	}
	return prev
}
