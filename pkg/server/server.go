package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	adminv1alpha1 "github.com/tacokumo/admin-api/pkg/apis/v1alpha1"
	adminv1alpha1generated "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/auth/oauth"
	"github.com/tacokumo/admin-api/pkg/auth/session"
	"github.com/tacokumo/admin-api/pkg/config"
	"github.com/tacokumo/admin-api/pkg/db/admindb"
	"github.com/tacokumo/admin-api/pkg/middleware"
	"github.com/tacokumo/admin-api/pkg/pg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Server struct {
	logger   *slog.Logger
	cfg      config.Config
	e        *echo.Echo
	cleanups []func(context.Context)
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Server, error) {
	s := &Server{
		logger: logger,
		cfg:    cfg,
		e:      echo.New(),
	}

	var cleanups []func(context.Context)

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to Redis")
	}
	cleanups = append(cleanups, func(ctx context.Context) {
		if err := redisClient.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close Redis connection", slog.String("error", err.Error()))
		}
	})

	// Initialize session stores
	sessionTTL := cfg.Auth.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = 24 * time.Hour
	}
	sessionStore := session.NewRedisStore(redisClient, sessionTTL)
	stateStore := session.NewRedisStore(redisClient, 10*time.Minute) // Short TTL for OAuth state

	// Initialize GitHub OAuth client
	githubClient := oauth.NewGitHubClient(
		cfg.Auth.GitHubClientID,
		cfg.Auth.GitHubClientSecret,
		cfg.Auth.CallbackURL,
		cfg.Auth.AllowedOrgs,
	)

	// Setup middleware
	s.e.Use(middleware.Logger(logger))
	corsConfig := setupCORSConfig(cfg)
	s.e.Use(echomiddleware.CORSWithConfig(corsConfig))
	sessionMiddleware := middleware.SessionMiddleware(logger, sessionStore)
	s.e.Use(sessionMiddleware)

	opts, otelCleanups, err := initAdminServerConfig(ctx, logger, cfg.Telemetry)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize admin server config")
	}
	cleanups = append(cleanups, otelCleanups...)

	adminDBConfig := pg.Config{
		Host:     cfg.AdminDBConfig.Host,
		Port:     cfg.AdminDBConfig.Port,
		User:     cfg.AdminDBConfig.User,
		Password: cfg.AdminDBConfig.Password,
		DBName:   cfg.AdminDBConfig.DBName,
	}
	pgxConfig, err := pgxpool.ParseConfig(adminDBConfig.DSN())
	if err != nil {
		return s, errors.Wrapf(err, "failed to parse pgx config")
	}

	// Only enable pgx tracing if telemetry is enabled
	if cfg.Telemetry.Enabled {
		pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	}

	p, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return s, errors.Wrapf(err, "failed to create pgx pool")
	}
	cleanups = append(cleanups, func(ctx context.Context) {
		p.Close()
	})

	retryCount := 1
	if cfg.AdminDBConfig.InitialConnRetry > 1 {
		retryCount = cfg.AdminDBConfig.InitialConnRetry
	}

	connected := false
	for i := 0; i < retryCount; i++ {
		err = p.Ping(ctx)
		if err == nil {
			connected = true
			break
		}
		logger.WarnContext(ctx, "failed to connect to admin db", slog.Int("retry_count", i+1), slog.Int("max_retry", retryCount), slog.String("error", err.Error()))
		time.Sleep(1 * time.Second)
	}
	if !connected {
		return s, errors.New("failed to connect to admin db")
	}

	// Only record pgx stats if telemetry is enabled
	if cfg.Telemetry.Enabled {
		if err := otelpgx.RecordStats(p); err != nil {
			return s, errors.Wrapf(err, "failed to record pgx stats")
		}
	}
	queries := admindb.New(p)

	// Create service with OAuth dependencies
	service := adminv1alpha1.NewService(
		logger,
		queries,
		githubClient,
		sessionStore,
		stateStore,
		cfg.Auth.FrontendURL,
		sessionTTL,
	)

	v1alpha1Server, err := adminv1alpha1generated.NewServer(
		service,
		service,
		opts...,
	)
	if err != nil {
		return s, errors.Wrapf(err, "failed to create v1alpha1 server")
	}

	// Register OAuth endpoints with Echo for proper redirect support
	s.e.GET("/v1alpha1/auth/login", createLoginHandler(logger, githubClient, stateStore))
	s.e.GET("/v1alpha1/auth/callback", createCallbackHandler(logger, githubClient, sessionStore, stateStore, cfg.Auth.FrontendURL, sessionTTL))

	v1alphaGroup := s.e.Group("/v1alpha1")
	v1alphaGroup.Any("/*", echo.WrapHandler(v1alpha1Server))

	s.cleanups = cleanups
	return s, nil
}

func createLoginHandler(
	logger *slog.Logger,
	githubClient *oauth.GitHubClient,
	stateStore session.Store,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Generate CSRF state
		state, err := session.GenerateSessionID()
		if err != nil {
			logger.ErrorContext(c.Request().Context(), "failed to generate state", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		// Store state with optional redirect_uri
		redirectURI := c.QueryParam("redirect_uri")
		stateSession := &session.Session{
			ID:        state,
			ExpiresAt: time.Now().Add(10 * time.Minute),
			CreatedAt: time.Now(),
		}
		if redirectURI != "" {
			stateSession.Name = redirectURI // Reuse Name field for redirect_uri
		}

		if err := stateStore.Create(c.Request().Context(), stateSession); err != nil {
			logger.ErrorContext(c.Request().Context(), "failed to store state", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		// Redirect to GitHub
		authURL := githubClient.GetAuthURL(state)
		return c.Redirect(http.StatusFound, authURL)
	}
}

func createCallbackHandler(
	logger *slog.Logger,
	githubClient *oauth.GitHubClient,
	sessionStore session.Store,
	stateStore session.Store,
	frontendURL string,
	sessionTTL time.Duration,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		code := c.QueryParam("code")
		state := c.QueryParam("state")

		if code == "" || state == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing code or state"})
		}

		// Validate state (CSRF protection)
		stateSession, err := stateStore.Get(ctx, state)
		if err != nil {
			logger.ErrorContext(ctx, "invalid state", slog.String("error", err.Error()))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid state"})
		}

		// Delete state after validation
		if err := stateStore.Delete(ctx, state); err != nil {
			logger.WarnContext(ctx, "failed to delete state", slog.String("error", err.Error()))
		}

		// Exchange code for token
		token, err := githubClient.ExchangeCode(ctx, code)
		if err != nil {
			logger.ErrorContext(ctx, "failed to exchange code", slog.String("error", err.Error()))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication failed"})
		}

		// Get user info
		ghUser, err := githubClient.GetUser(ctx, token)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get user info", slog.String("error", err.Error()))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "failed to get user info"})
		}

		// Validate org membership
		orgs, err := githubClient.GetUserOrgs(ctx, token)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get user orgs", slog.String("error", err.Error()))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "failed to get user orgs"})
		}

		if !githubClient.ValidateOrgMembership(orgs) {
			logger.WarnContext(ctx, "user not in allowed org", slog.String("username", ghUser.Login))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not authorized: not a member of allowed organizations"})
		}

		// Get team memberships
		teams, err := githubClient.GetTeamMemberships(ctx, token)
		if err != nil {
			logger.WarnContext(ctx, "failed to get team memberships", slog.String("error", err.Error()))
			teams = []oauth.TeamMembership{}
		}

		// Generate session ID
		sessionID, err := session.GenerateSessionID()
		if err != nil {
			logger.ErrorContext(ctx, "failed to generate session ID", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		// Create session
		sess := &session.Session{
			ID:             sessionID,
			UserID:         fmt.Sprintf("%d", ghUser.ID),
			GitHubUserID:   ghUser.ID,
			GitHubUsername: ghUser.Login,
			Email:          ghUser.Email,
			Name:           ghUser.Name,
			AvatarURL:      ghUser.AvatarURL,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			TeamMemberships: lo.Map(teams, func(tm oauth.TeamMembership, _ int) session.TeamMembership {
				return session.TeamMembership{
					OrgName:  tm.OrgName,
					TeamName: tm.TeamName,
					Role:     tm.Role,
				}
			}),
			ExpiresAt: time.Now().Add(sessionTTL),
			CreatedAt: time.Now(),
		}

		if err := sessionStore.Create(ctx, sess); err != nil {
			logger.ErrorContext(ctx, "failed to create session", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		// Set session cookie
		c.SetCookie(&http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(sessionTTL.Seconds()),
		})

		// Redirect to frontend
		redirectURL := frontendURL
		if stateSession.Name != "" {
			redirectURL = stateSession.Name // Use stored redirect_uri
		}
		redirectURL = fmt.Sprintf("%s?token=%s&state=%s", redirectURL, sessionID, state)

		return c.Redirect(http.StatusFound, redirectURL)
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Start server
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go startAPIServer(ctx, s.logger, s.e, s.cfg, wg)

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.e.Shutdown(ctx); err != nil {
		return errors.Wrapf(err, "failed to shutdown server")
	}

	wg.Wait()

	for _, cleanup := range s.cleanups {
		cleanup(ctx)
	}
	return nil
}

func initAdminServerConfig(
	ctx context.Context,
	logger *slog.Logger,
	telemetryCfg config.TelemetryConfig,
) ([]adminv1alpha1generated.ServerOption, []func(context.Context), error) {
	var opts []adminv1alpha1generated.ServerOption
	var cleanups []func(context.Context)

	// If telemetry is disabled, return empty options
	if !telemetryCfg.Enabled {
		logger.InfoContext(ctx, "telemetry is disabled")
		return opts, cleanups, nil
	}

	res, err := newResource()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create resource")
	}

	// Configure OTLP endpoint and timeout if specified
	var traceExporterOpts []otlptracegrpc.Option
	var metricExporterOpts []otlpmetricgrpc.Option

	if telemetryCfg.OTLPEndpoint != "" {
		traceExporterOpts = append(traceExporterOpts, otlptracegrpc.WithEndpoint(telemetryCfg.OTLPEndpoint))
		metricExporterOpts = append(metricExporterOpts, otlpmetricgrpc.WithEndpoint(telemetryCfg.OTLPEndpoint))
	}

	if telemetryCfg.Timeout > 0 {
		traceExporterOpts = append(traceExporterOpts, otlptracegrpc.WithTimeout(telemetryCfg.Timeout))
		metricExporterOpts = append(metricExporterOpts, otlpmetricgrpc.WithTimeout(telemetryCfg.Timeout))
	}

	// STEP1: TracerProvider
	traceExporter, err := otlptracegrpc.New(ctx, traceExporterOpts...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create trace exporter")
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSyncer(traceExporter),
	)
	cleanups = append(cleanups, func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown TracerProvider", slog.String("error", err.Error()))
		}
	})

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	opts = append(opts, adminv1alpha1generated.WithTracerProvider(tp))

	// STEP2: MeterProvider
	meterExporter, err := otlpmetricgrpc.New(ctx, metricExporterOpts...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create meter exporter")
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(meterExporter)),
	)
	cleanups = append(cleanups, func(ctx context.Context) {
		if err := mp.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown MeterProvider", slog.String("error", err.Error()))
		}
	})
	otel.SetMeterProvider(mp)
	opts = append(opts, adminv1alpha1generated.WithMeterProvider(mp))

	return opts, cleanups, nil
}

func newResource() (*resource.Resource, error) {
	r, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("tacokumo-admin"),
			semconv.ServiceVersion("0.1.0"),
		))
	if err != nil {
		return nil, err
	}
	return r, nil
}

func startAPIServer(ctx context.Context, logger *slog.Logger, e *echo.Echo, cfg config.Config, wg *sync.WaitGroup) {
	defer wg.Done()
	addr := net.JoinHostPort(cfg.Addr, cfg.Port)

	if !cfg.TLS.Enabled {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.ErrorContext(ctx, "shutting down the server", slog.String("addr", addr), slog.String("port", cfg.Port), slog.String("error", err.Error()))
		}
		return
	}

	// Setup TLS with mTLS
	tlsConfig, err := setupTLSConfig(cfg.TLS)
	if err != nil {
		logger.ErrorContext(ctx, "failed to setup TLS config", slog.String("error", err.Error()))
		return
	}

	server := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
		Handler:   e,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		logger.ErrorContext(ctx, "shutting down the server", slog.String("addr", addr), slog.String("port", cfg.Port), slog.String("error", err.Error()))
	}
}

func setupTLSConfig(tlsCfg config.TLSConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load X509 key pair")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func setupCORSConfig(cfg config.Config) echomiddleware.CORSConfig {
	allowOrigins := strings.Split(cfg.CORS.AllowOrigins, ",")
	allowMethods := strings.Split(cfg.CORS.AllowMethods, ",")
	allowHeaders := strings.Split(cfg.CORS.AllowHeaders, ",")
	exposeHeaders := strings.Split(cfg.CORS.ExposeHeaders, ",")

	corsConfig := echomiddleware.CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	return corsConfig
}
