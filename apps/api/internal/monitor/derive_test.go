package monitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var defaultT = Thresholds{DegradedMs: 500, N: 3, K: 2}

func up(latency int) Sample { return Sample{Responded: true, OK: true, LatencyMs: latency} }
func fail() Sample          { return Sample{Responded: false, OK: false} }
func wrongCode() Sample     { return Sample{Responded: true, OK: false, LatencyMs: 10} }
func slow() Sample          { return Sample{Responded: true, OK: true, LatencyMs: 1200} }

// run feeds a sequence of samples through Derive, threading state, and returns
// the final state plus the list of states after each step.
func run(start Status, samples []Sample) (Status, []Status, []bool) {
	st := start
	var cf, cs int
	states := []Status{}
	changed := []bool{}
	for _, s := range samples {
		r := Derive(DeriveInput{Prev: st, ConsecFailures: cf, ConsecSuccesses: cs, Sample: s, T: defaultT})
		st, cf, cs = r.State, r.ConsecFailures, r.ConsecSuccesses
		states = append(states, r.State)
		changed = append(changed, r.Changed)
	}
	return st, states, changed
}

func TestFirstSuccessGoesUp(t *testing.T) {
	r := Derive(DeriveInput{Prev: StatusUnknown, Sample: up(50), T: defaultT})
	require.Equal(t, StatusUp, r.State)
	require.True(t, r.Changed)
	require.Equal(t, StatusUp, r.Outcome)
}

func TestAntiFlapDownNeedsNFailures(t *testing.T) {
	// From up, two failures hold "up", the third declares "down".
	final, states, changed := run(StatusUp, []Sample{fail(), fail(), fail()})
	require.Equal(t, []Status{StatusUp, StatusUp, StatusDown}, states)
	require.Equal(t, []bool{false, false, true}, changed)
	require.Equal(t, StatusDown, final)
}

func TestWrongCodeCountsAsFailure(t *testing.T) {
	// HTTP 200-but-wrong-code path: Responded=true, OK=false → failure.
	final, _, _ := run(StatusUp, []Sample{wrongCode(), wrongCode(), wrongCode()})
	require.Equal(t, StatusDown, final)
}

func TestRecoveryNeedsKSuccesses(t *testing.T) {
	// From down, one success keeps down; the second recovers to up.
	final, states, changed := run(StatusDown, []Sample{up(40), up(40)})
	require.Equal(t, []Status{StatusDown, StatusUp}, states)
	require.Equal(t, []bool{false, true}, changed)
	require.Equal(t, StatusUp, final)
}

func TestDegradedOnSlowSuccess(t *testing.T) {
	r := Derive(DeriveInput{Prev: StatusUp, Sample: slow(), T: defaultT})
	require.Equal(t, StatusDegraded, r.State)
	require.True(t, r.Changed)
	require.Equal(t, StatusDegraded, r.Outcome)
}

func TestUpDegradedEdgeIsImmediate(t *testing.T) {
	final, states, _ := run(StatusUp, []Sample{slow(), up(20), slow()})
	require.Equal(t, []Status{StatusDegraded, StatusUp, StatusDegraded}, states)
	require.Equal(t, StatusDegraded, final)
}

func TestFailureResetsSuccessStreakAndViceVersa(t *testing.T) {
	// up, fail (1/3, hold up), success resets failure streak, so it takes a
	// fresh run of 3 failures to go down.
	final, states, _ := run(StatusUp, []Sample{fail(), up(30), fail(), fail()})
	require.Equal(t, []Status{StatusUp, StatusUp, StatusUp, StatusUp}, states)
	require.Equal(t, StatusUp, final)
}

func TestRecoveryInterruptedByFailure(t *testing.T) {
	// down, one success (recovering 1/2), a failure resets and keeps down.
	final, states, changed := run(StatusDown, []Sample{up(30), fail(), up(30), up(30)})
	require.Equal(t, []Status{StatusDown, StatusDown, StatusDown, StatusUp}, states)
	require.Equal(t, []bool{false, false, false, true}, changed)
	require.Equal(t, StatusUp, final)
}

func TestUnknownHoldsUntilNFailures(t *testing.T) {
	// A fresh service that never responded should not alert until N failures.
	final, states, changed := run(StatusUnknown, []Sample{fail(), fail()})
	require.Equal(t, []Status{StatusUnknown, StatusUnknown}, states)
	require.Equal(t, []bool{false, false}, changed)
	require.Equal(t, StatusUnknown, final)
}

func TestRecoverToDegraded(t *testing.T) {
	// Recovery from down can land directly in degraded if responses are slow.
	final, _, _ := run(StatusDown, []Sample{slow(), slow()})
	require.Equal(t, StatusDegraded, final)
}
