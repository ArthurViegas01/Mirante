package monitor

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func discardLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func openTestDB(t *testing.T) *idb.DB {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	_, err = database.ExecContext(ctx, `INSERT INTO projects (id, nome) VALUES ('proj1', 'Proj')`)
	require.NoError(t, err)
	return database
}

type scriptChecker struct {
	samples []Sample
	i       int
}

func (c *scriptChecker) Check(_ context.Context, _ *Service) Sample {
	s := c.samples[c.i%len(c.samples)]
	c.i++
	return s
}

type captureSink struct{ count int }

func (s *captureSink) Emit(_ string, _ int64, _ string, _ []byte) { s.count++ }

func newService(id string) *Service {
	return &Service{
		ID: ServiceID(id), ProjectID: "proj1", Nome: "API", Kind: KindHTTP,
		Target: "http://example.test", ExpectedStatus: "2xx",
		DegradedThresholdMs: 500, TimeoutMs: 5000, IntervalSeconds: 60,
		AntiFlapN: 2, RecoveryK: 1, Enabled: true, CurrentStatus: StatusUnknown,
	}
}

func TestEngineTransitionPersistsAtomically(t *testing.T) {
	ctx := context.Background()
	repo := NewSQLiteRepo(openTestDB(t))
	require.NoError(t, repo.CreateService(ctx, newService("svc1")))

	sink := &captureSink{}
	eng := NewEngine(repo, &scriptChecker{samples: []Sample{fail(), fail()}}, NewNotifier(discardLog()), sink, discardLog())

	svc, err := repo.GetService(ctx, "svc1")
	require.NoError(t, err)

	// First failure holds "unknown" (anti-flap), no transition.
	require.NoError(t, eng.RunCheck(ctx, svc))
	require.Equal(t, StatusUnknown, svc.CurrentStatus)
	require.Equal(t, 0, sink.count)

	// Second failure crosses N=2 → down, a transition.
	require.NoError(t, eng.RunCheck(ctx, svc))
	require.Equal(t, StatusDown, svc.CurrentStatus)
	require.Equal(t, 1, sink.count)

	// One alert persisted, danger severity.
	alerts, err := repo.ListAlerts(ctx, 10, false)
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	require.Equal(t, "danger", alerts[0].Severity)
	require.Equal(t, StatusDown, alerts[0].ToStatus)

	// Two check_results recorded; one event in the outbox.
	checks, err := repo.ListChecks(ctx, "svc1", 10)
	require.NoError(t, err)
	require.Len(t, checks, 2)

	events, err := repo.EventsAfter(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "monitor.transition", events[0].Type)

	// Uptime over 24h reflects two down samples.
	up, err := repo.Uptime(ctx, "svc1", 24)
	require.NoError(t, err)
	require.Equal(t, 2, up.Samples)
	require.Equal(t, 0.0, up.UpRatio)
}

func TestEngineRecoveryEmitsTwoTransitions(t *testing.T) {
	ctx := context.Background()
	repo := NewSQLiteRepo(openTestDB(t))
	require.NoError(t, repo.CreateService(ctx, newService("svc2")))

	sink := &captureSink{}
	// fail, fail → down; up (K=1) → recovered.
	eng := NewEngine(repo, &scriptChecker{samples: []Sample{fail(), fail(), up(30)}}, NewNotifier(discardLog()), sink, discardLog())
	svc, _ := repo.GetService(ctx, "svc2")
	for i := 0; i < 3; i++ {
		require.NoError(t, eng.RunCheck(ctx, svc))
	}
	require.Equal(t, StatusUp, svc.CurrentStatus)
	require.Equal(t, 2, sink.count) // down + recovery
}

func TestSchedulerStartStopNoLeak(t *testing.T) {
	// VerifyNone runs last (LIFO); the DB is closed before it so the sql.DB
	// cleaner/opener goroutines are gone and not mistaken for a leak.
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	require.NoError(t, migrate.Up(database.DB))
	_, err = database.ExecContext(ctx, `INSERT INTO projects (id, nome) VALUES ('proj1', 'Proj')`)
	require.NoError(t, err)

	repo := NewSQLiteRepo(database)
	svc := newService("s1")
	svc.Kind = KindTCP
	svc.Target = "127.0.0.1:1" // connection refused → fast, deterministic failure
	svc.IntervalSeconds = 5
	svc.TimeoutMs = 1000 // must stay < interval_seconds*1000
	require.NoError(t, repo.CreateService(ctx, svc))

	eng := NewEngine(repo, NewChecker(), NewNotifier(discardLog()), &captureSink{}, discardLog())
	sched := NewScheduler(repo, eng, discardLog(), 4)

	sched.Start(ctx)
	time.Sleep(150 * time.Millisecond) // let the initial check run
	sched.Stop()                       // must return without leaking goroutines
}
