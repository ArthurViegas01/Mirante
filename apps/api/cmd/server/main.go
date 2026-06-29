// Command server is the Mirante API entrypoint (composition root).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/lumni/mirante/internal/applications"
	"github.com/lumni/mirante/internal/cv"
	"github.com/lumni/mirante/internal/intake"
	"github.com/lumni/mirante/internal/jobs"
	"github.com/lumni/mirante/internal/llm"
	"github.com/lumni/mirante/internal/monitor"
	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/config"
	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/httpserver"
	"github.com/lumni/mirante/internal/platform/httpx"
	"github.com/lumni/mirante/internal/platform/logging"
	"github.com/lumni/mirante/internal/platform/mailer"
	"github.com/lumni/mirante/internal/platform/migrate"
	"github.com/lumni/mirante/internal/platform/otel"
	"github.com/lumni/mirante/internal/platform/ratelimit"
	"github.com/lumni/mirante/internal/platform/sse"
	"github.com/lumni/mirante/internal/projects"
	"github.com/lumni/mirante/internal/subscriptions"
	"github.com/lumni/mirante/internal/tasks"
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
	resetMailer, mailerOK := buildMailer(cfg, log)
	authSvc.WithMailer(resetMailer, cfg.WebOrigin, cfg.PasswordResetTTL)
	if !mailerOK {
		log.Info("password reset e-mail disabled — reset links are logged to the server")
	}
	if err := authSvc.Bootstrap(ctx, cfg.OwnerEmail, cfg.OwnerPassword, cfg.OwnerHash); err != nil {
		return err
	}
	if needs, err := authSvc.NeedsSetup(ctx); err == nil && needs {
		log.Info("no owner configured — the instance will be claimed via first-run signup")
	}

	// Trust the real-client-IP header only when a trusted proxy fronts the app
	// (TRUSTED_PROXY=true on Railway/Fly). Empty header = ignore proxy headers so a
	// directly-exposed instance can't be spoofed past the per-IP limits (F4).
	trustedHeader := ""
	if cfg.TrustedProxy {
		trustedHeader = cfg.RealIPHeader
	} else if cfg.IsProd() {
		log.Warn("TRUSTED_PROXY is false in production — per-IP rate limits will key on the " +
			"edge proxy address (one shared bucket for all clients). Set TRUSTED_PROXY=true behind Railway/Fly.")
	}

	authH := httpserver.NewAuthHandlers(authSvc, httpserver.AuthConfig{
		CookieName:    cfg.SessionCookie,
		Secure:        cfg.IsProd(),
		AllowedOrigin: cfg.WebOrigin,
		TrustedHeader: trustedHeader,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", httpserver.Healthz)
	authH.RegisterRoutes(mux)

	// Project GitHub import uses the SSRF-guarded external policy (ADR-0003):
	// private IPs blocked. GitHub's API requires a User-Agent.
	githubFetcher := httpx.NewFetcher(httpx.Policy{
		AllowPrivateIPs: false,
		MaxBodyBytes:    256 << 10,
		UserAgent:       "Mirante/1.0 (+https://github.com/lumni/mirante)",
	})
	projectsSvc := projects.NewService(projects.NewSQLiteRepo(database), githubFetcher)
	projects.RegisterRoutes(mux, authH.Protect, projectsSvc)

	tasksSvc := tasks.NewService(tasks.NewSQLiteRepo(database))
	tasks.RegisterRoutes(mux, authH.Protect, tasksSvc)

	subsSvc := subscriptions.NewService(subscriptions.NewSQLiteRepo(database))
	subscriptions.RegisterRoutes(mux, authH.Protect, subsSvc)

	// LLM gateway (ADR-0004). Absent key → unavailable client; features degrade.
	var llmClient *llm.Client
	if provider, ok := llm.NewProvider(cfg.LLMProvider, cfg.LLMModel, cfg.LLMAPIKey); ok {
		llmClient = llm.NewClient(provider, llm.NewSQLiteLedger(database), llm.NewRouteLimiter(cfg.LLMRatePerMinute))
		log.Info("llm enabled", "provider", provider.Name(), "model", provider.Model())
	} else {
		llmClient = llm.NewClient(nil, nil, nil)
		log.Info("llm disabled (no API key)")
	}

	// Job-link import uses the SSRF-guarded JobLink policy (ADR-0003): private IPs
	// blocked, with a browser-like UA to read public postings (e.g. LinkedIn).
	jobLinkFetcher := httpx.NewFetcher(httpx.Policy{
		AllowPrivateIPs: false,
		MaxBodyBytes:    1 << 20,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})
	jobsSvc := jobs.NewService(jobs.NewSQLiteRepo(database), llmClient, jobLinkFetcher)
	jobs.RegisterRoutes(mux, authH.Protect, jobsSvc)

	cvSvc := cv.NewService(cv.NewSQLiteRepo(database), llmClient)
	cv.RegisterRoutes(mux, authH.Protect, cvSvc)

	applicationsSvc := applications.NewService(applications.NewSQLiteRepo(database))
	applications.RegisterRoutes(mux, authH.Protect, applicationsSvc)

	// Intake: the freelance-opportunity funnel (99Freelas). The service + routes are
	// always mounted (so the funnel UI works); the IMAP poller that fills it starts
	// only when creds are configured. Scoring reuses the owner's CV skills via a port
	// (ADR-0001); ingested items are attributed to the admin.
	intakeSvc := intake.NewService(
		intake.NewSQLiteRepo(database),
		func(ctx context.Context) ([]string, error) {
			p, err := cvSvc.GetProfile(ctx)
			if err != nil {
				return nil, err
			}
			return p.Skills, nil
		},
		cfg.IntakeMinScore,
	)
	intake.RegisterRoutes(mux, authH.Protect, intakeSvc)

	var intakeStopped <-chan struct{}
	if cfg.IntakeEnabled() {
		source := intake.NewIMAPSource(intake.IMAPConfig{
			Host:     cfg.IntakeIMAPHost,
			Port:     cfg.IntakeIMAPPort,
			Username: cfg.IntakeIMAPUsername,
			Password: cfg.IntakeIMAPPassword,
			Mailbox:  cfg.IntakeIMAPMailbox,
			From:     cfg.IntakeIMAPFrom,
		})
		userStore := auth.NewUserStore(database.DB)
		owner := func(ctx context.Context) (string, error) {
			// The polled inbox belongs to one Mirante account: the one whose e-mail is
			// the IMAP login. Attribute ingested items there, falling back to the admin
			// (covers a single-user instance where the two coincide, or a misconfigured
			// e-mail).
			if email := strings.ToLower(strings.TrimSpace(cfg.IntakeIMAPUsername)); email != "" {
				if u, err := userStore.GetByEmail(ctx, email); err == nil {
					return u.ID, nil
				}
			}
			u, err := userStore.Admin(ctx)
			if errors.Is(err, auth.ErrUserNotFound) {
				return "", nil // unclaimed instance — idle until first signup
			}
			if err != nil {
				return "", err
			}
			return u.ID, nil
		}
		intakeStopped = intake.NewRunner(source, intakeSvc, owner, cfg.IntakePollInterval, log).Start(ctx)
		log.Info("intake poller started",
			"mailbox", cfg.IntakeIMAPMailbox, "from", cfg.IntakeIMAPFrom, "interval", cfg.IntakePollInterval)
	}

	monitorRepo := monitor.NewSQLiteRepo(database)
	monitorMgr := monitor.NewManager(monitorRepo)
	hub := sse.NewHub(func(ctx context.Context, afterID int64, limit int) ([]sse.Event, error) {
		evs, err := monitorRepo.EventsAfter(ctx, afterID, limit)
		if err != nil {
			return nil, err
		}
		out := make([]sse.Event, len(evs))
		for i, e := range evs {
			out[i] = sse.Event{ID: e.ID, Type: e.Type, Data: e.Data}
		}
		return out, nil
	})
	// Optional external alert delivery (F5): an owner-configured webhook receives
	// each monitor transition. Absent/invalid URL → no channel (in-app only).
	var alertChannels []monitor.AlertChannel
	if cfg.AlertWebhookURL != "" {
		if ch, err := monitor.NewWebhookChannel(cfg.AlertWebhookURL, nil); err != nil {
			log.Warn("invalid ALERT_WEBHOOK_URL — alert webhook disabled", "err", err)
		} else {
			alertChannels = append(alertChannels, ch)
			log.Info("alert webhook enabled")
		}
	}
	monitorEngine := monitor.NewEngine(monitorRepo, monitor.NewChecker(), monitor.NewNotifier(log, alertChannels...), hub, log)
	monitorSched := monitor.NewScheduler(monitorRepo, monitorEngine, log, 8)
	monitorMgr.SetReconciler(monitorSched)
	monitor.RegisterRoutes(mux, authH.Protect, monitorMgr)
	mux.Handle("GET /api/stream/monitor", authH.RequireAuth(hub))

	ipLimiter := ratelimit.New(240, time.Minute)
	handler := httpserver.Chain(mux,
		httpserver.RequestID(),
		httpserver.Recover(log),
		httpserver.SecurityHeaders(cfg.IsProd()),
		httpserver.CORS(cfg.WebOrigin),
		httpserver.RateLimit(ipLimiter, trustedHeader),
	)

	// Outermost: an OTel server span per request (extracts trace context first;
	// a no-op when no exporter is configured). Method/status land as attributes.
	traced := otelhttp.NewHandler(handler, "http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return "HTTP " + r.Method }))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           traced,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,  // headers + body (slow-body cap)
		IdleTimeout:       120 * time.Second, // idle keep-alive cap
		// WriteTimeout is omitted on purpose: /api/stream/monitor is a long-lived SSE
		// response and a write deadline would kill the stream.
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

	// Background: roll up old monitor checks into hourly buckets and prune the raw
	// rows, keeping check_results bounded while long-window uptime stays computable
	// (F4). Runs once on boot, then hourly until the context is cancelled.
	compactDone := make(chan struct{})
	go func() {
		defer close(compactDone)
		compact := func() {
			if n, err := monitorMgr.Compact(ctx, cfg.MonitorRetention); err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Warn("monitor compaction failed", "err", err)
				}
			} else if n > 0 {
				log.Info("compacted monitor checks", "rows", n)
			}
		}
		compact()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				compact()
			}
		}
	}()

	monitorSched.Start(ctx)
	log.Info("monitor scheduler started")

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
		monitorSched.Stop()
		<-sweepDone
		<-compactDone
		if intakeStopped != nil {
			<-intakeStopped
		}
		return err
	}
}

// buildMailer picks the e-mail transport: the Resend HTTP API when
// RESEND_API_KEY is set (preferred for production deliverability), else SMTP when
// SMTP_HOST is set (Mailpit in dev), else none — in which case the reset flow
// logs the link to the server. An invalid config disables e-mail rather than
// blocking boot.
func buildMailer(cfg config.Config, log *slog.Logger) (auth.Mailer, bool) {
	if cfg.ResendAPIKey != "" {
		m, err := mailer.NewResend(cfg.ResendAPIKey, cfg.MailFrom)
		if err != nil {
			log.Warn("invalid Resend config — password reset e-mail disabled", "err", err)
			return nil, false
		}
		log.Info("password reset e-mail enabled", "provider", "resend", "from", cfg.MailFrom)
		return m, true
	}
	if cfg.SMTPHost != "" {
		m, err := mailer.NewSMTP(mailer.Config{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.MailFrom,
		})
		if err != nil {
			log.Warn("invalid SMTP config — password reset e-mail disabled", "err", err)
			return nil, false
		}
		log.Info("password reset e-mail enabled", "provider", "smtp", "smtp_host", cfg.SMTPHost)
		return m, true
	}
	return nil, false
}
