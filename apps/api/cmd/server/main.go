// Command server is the Mirante API entrypoint (composition root).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/config"
	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/httpserver"
	"github.com/lumni/mirante/internal/platform/logging"
	"github.com/lumni/mirante/internal/platform/migrate"
	"github.com/lumni/mirante/internal/platform/otel"
	"github.com/lumni/mirante/internal/platform/ratelimit"
	"github.com/lumni/mirante/internal/projects"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logging.New(cfg.AppEnv)
	slog.SetDefault(log)

	_, shutdownOtel := otel.Init(cfg.OtelService, cfg.OtelEndpoint)
	defer func() { _ = shutdownOtel(context.Background()) }()

	// Root context cancelled on SIGINT/SIGTERM; drives background workers.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(ctx, cfg.DatabaseURL, cfg.DatabaseToken)
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	if err := migrate.Up(database.DB); err != nil {
		return err
	}
	log.Info("migrations applied")

	authSvc := auth.NewService(database.DB, cfg.SessionTTL)
	if err := authSvc.Bootstrap(ctx, cfg.OwnerEmail, cfg.OwnerPassword, cfg.OwnerHash); err != nil {
		return err
	}

	authH := httpserver.NewAuthHandlers(authSvc, httpserver.AuthConfig{
		CookieName:    cfg.SessionCookie,
		Secure:        cfg.IsProd(),
		AllowedOrigin: cfg.WebOrigin,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", httpserver.Healthz)
	authH.RegisterRoutes(mux)

	projectsSvc := projects.NewService(projects.NewSQLiteRepo(database))
	projects.RegisterRoutes(mux, authH.Protect, projectsSvc)

	ipLimiter := ratelimit.New(240, time.Minute)
	handler := httpserver.Chain(mux,
		httpserver.RequestID(),
		httpserver.Recover(log),
		httpserver.SecurityHeaders(cfg.IsProd()),
		httpserver.CORS(cfg.WebOrigin),
		httpserver.RateLimit(ipLimiter),
	)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Background: sweep expired sessions hourly until the context is cancelled.
	sweepDone := make(chan struct{})
	go func() {
		defer close(sweepDone)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := authSvc.SweepExpired(ctx); err != nil {
					log.Warn("session sweep failed", "err", err)
				} else if n > 0 {
					log.Info("swept expired sessions", "count", n)
				}
			}
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := srv.Shutdown(shutdownCtx)
		<-sweepDone
		return err
	}
}
