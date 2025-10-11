package cmd

import (
	"context"
	"crypto/tls"
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
	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/a0/management"
	adminv1alpha1 "github.com/tacokumo/admin-api/pkg/apis/v1alpha1"
	adminv1alpha1generated "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
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

func New(logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:           "api",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadFromEnv()
			if err != nil {
				return errors.WithStack(err)
			}

			a0ManagementClient, err := management.NewDefaultClient(cmd.Context(), cfg.Auth.Auth0Domain, cfg.Auth.Auth0ClientID, cfg.Auth.Auth0ClientSecret)
			if err != nil {
				return errors.WithStack(err)
			}
			// Setup
			e := echo.New()
			e.Use(middleware.Logger(logger))
			corsConfig := setupCORSConfig(cfg)
			e.Use(echomiddleware.CORSWithConfig(corsConfig))
			jwtValidateMiddleware, err := middleware.JWTMiddleware(cfg.Auth.Auth0Domain, 5*time.Minute, []string{cfg.Auth.Auth0Audience})
			if err != nil {
				return errors.WithStack(err)
			}
			e.Use(jwtValidateMiddleware)
			e.Use(middleware.AuthZ(logger, a0ManagementClient, middleware.WithExclusionURLPrefixes([]string{
				"/v1alpha1/health/liveness",
				"/v1alpha1/health/readiness",
			})))

			opts, cleanups, err := initAdminServerConfig(cmd.Context(), logger)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, cleanup := range cleanups {
				defer cleanup(cmd.Context())
			}

			adminDBConfig := pg.Config{
				Host:     cfg.AdminDBConfig.Host,
				Port:     cfg.AdminDBConfig.Port,
				User:     cfg.AdminDBConfig.User,
				Password: cfg.AdminDBConfig.Password,
				DBName:   cfg.AdminDBConfig.DBName,
			}
			pgxConfig, err := pgxpool.ParseConfig(adminDBConfig.DSN())
			if err != nil {
				return errors.WithStack(err)
			}

			pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()

			p, err := pgxpool.NewWithConfig(cmd.Context(), pgxConfig)
			if err != nil {
				return errors.WithStack(err)
			}
			defer p.Close()

			retryCount := 1
			if cfg.AdminDBConfig.InitialConnRetry > 1 {
				retryCount = cfg.AdminDBConfig.InitialConnRetry
			}

			connected := false
			for i := 0; i < retryCount; i++ {
				err = p.Ping(cmd.Context())
				if err == nil {
					connected = true
					break
				}
				logger.WarnContext(cmd.Context(), "failed to connect to admin db", slog.Int("retry_count", i+1), slog.Int("max_retry", retryCount), slog.String("error", err.Error()))
				time.Sleep(1 * time.Second)
			}
			if !connected {
				return errors.New("failed to connect to admin db")
			}

			if err := otelpgx.RecordStats(p); err != nil {
				return errors.WithStack(err)
			}
			queries := admindb.New(p)

			v1alpha1Service, err := adminv1alpha1generated.NewServer(
				adminv1alpha1.NewService(logger, queries),
				opts...,
			)
			if err != nil {
				return errors.WithStack(err)
			}
			v1alphaGroup := e.Group("/v1alpha1")
			v1alphaGroup.Any("/*", echo.WrapHandler(v1alpha1Service))

			// Start server
			wg := new(sync.WaitGroup)

			wg.Add(1)
			go startAPIServer(cmd.Context(), logger, e, cfg, wg)

			// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
			<-cmd.Context().Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := e.Shutdown(ctx); err != nil {
				return errors.WithStack(err)
			}

			wg.Wait()
			return nil
		},
	}
}

func initAdminServerConfig(
	ctx context.Context,
	logger *slog.Logger,
) ([]adminv1alpha1generated.ServerOption, []func(context.Context), error) {
	var opts []adminv1alpha1generated.ServerOption
	var cleanups []func(context.Context)

	res, err := newResource()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// STEP1: TracerProvider
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
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
	meterExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
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
		return nil, errors.WithStack(err)
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
