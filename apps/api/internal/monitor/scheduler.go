package monitor

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type serviceLoop struct {
	cancel    context.CancelFunc
	updatedAt time.Time
}

// Scheduler runs each enabled service's checks on its own goroutine, caps global
// concurrency with a semaphore, and periodically reconciles with the database to
// pick up added/removed/edited services (a per-service CancelFunc registry).
type Scheduler struct {
	repo          Repository
	engine        *Engine
	log           *slog.Logger
	reconcileEach time.Duration

	sem chan struct{}

	mu      sync.Mutex
	running map[ServiceID]*serviceLoop
	wg      sync.WaitGroup
	rootCtx context.Context
	cancel  context.CancelFunc
}

// NewScheduler builds a scheduler capping concurrent checks at maxConcurrent.
func NewScheduler(repo Repository, engine *Engine, log *slog.Logger, maxConcurrent int) *Scheduler {
	if maxConcurrent < 1 {
		maxConcurrent = 8
	}
	return &Scheduler{
		repo:          repo,
		engine:        engine,
		log:           log,
		reconcileEach: 15 * time.Second,
		sem:           make(chan struct{}, maxConcurrent),
		running:       make(map[ServiceID]*serviceLoop),
	}
}

// Start launches the scheduler; it returns immediately.
func (s *Scheduler) Start(ctx context.Context) {
	s.rootCtx, s.cancel = context.WithCancel(ctx)
	s.reconcile()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.reconcileEach)
		defer ticker.Stop()
		for {
			select {
			case <-s.rootCtx.Done():
				return
			case <-ticker.C:
				s.reconcile()
			}
		}
	}()
}

// Stop cancels the scheduler's context (which cascades to the reconcile loop and
// every service loop) and waits for all goroutines to exit (no leaks).
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) reconcile() {
	services, err := s.repo.ListEnabledServices(s.rootCtx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			s.log.Warn("monitor reconcile: list services", "err", err)
		}
		return
	}
	want := make(map[ServiceID]*Service, len(services))
	for _, svc := range services {
		want[svc.ID] = svc
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop loops for removed/disabled services, or those whose config changed
	// (config edits bump updated_at; routine checks do not).
	for id, loop := range s.running {
		svc, ok := want[id]
		if !ok || !svc.UpdatedAt.Equal(loop.updatedAt) {
			loop.cancel()
			delete(s.running, id)
		}
	}
	// Start loops for new or restarted services.
	for id, svc := range want {
		if _, ok := s.running[id]; ok {
			continue
		}
		lctx, cancel := context.WithCancel(s.rootCtx)
		s.running[id] = &serviceLoop{cancel: cancel, updatedAt: svc.UpdatedAt}
		s.wg.Add(1)
		go s.loop(lctx, svc)
	}
}

// loop runs checks for a single service sequentially (single-flight per
// service: a slow check delays the next tick rather than overlapping).
func (s *Scheduler) loop(ctx context.Context, svc *Service) {
	defer s.wg.Done()
	ticker := time.NewTicker(time.Duration(svc.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	s.runOne(ctx, svc)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOne(ctx, svc)
		}
	}
}

func (s *Scheduler) runOne(ctx context.Context, svc *Service) {
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return
	}
	defer func() { <-s.sem }()

	if err := s.engine.RunCheck(ctx, svc); err != nil && !errors.Is(err, context.Canceled) {
		s.log.Warn("monitor check failed", "service", svc.ID, "err", err)
	}
}
