package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/ces1231/blue-ledger-api/internal/dues"
	"github.com/ces1231/blue-ledger-api/internal/events"
	"github.com/ces1231/blue-ledger-api/internal/members"
	"github.com/ces1231/blue-ledger-api/internal/notifications"
	"github.com/ces1231/blue-ledger-api/internal/xp"
	"github.com/ces1231/blue-ledger-api/pkg/config"
	appdb "github.com/ces1231/blue-ledger-api/pkg/db"
	appredis "github.com/ces1231/blue-ledger-api/pkg/redis"
	"github.com/ces1231/blue-ledger-api/pkg/email"
	"github.com/ces1231/blue-ledger-api/pkg/logger"
	"github.com/ces1231/blue-ledger-api/pkg/validator"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load .env if present (ignored in production where env vars are injected directly)
	_ = godotenv.Load()

	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	ctx := context.Background()

	// ── Database ──────────────────────────────────────────────────────────────
	pool, err := appdb.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("database connected")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient, err := appredis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("failed to connect to redis — continuing without cache")
		redisClient = nil
	} else {
		log.Info().Msg("redis connected")
	}
	_ = redisClient // used by rate limiter in future tasks

	// ── Run Migrations ────────────────────────────────────────────────────────
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("failed to run database migrations")
	}
	log.Info().Msg("migrations applied")

	// ── Services ──────────────────────────────────────────────────────────────
	tokenManager, err := auth.NewTokenManager(
		cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath,
		cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load JWT keys — generate RSA keys with: make gen-keys")
	}

	emailClient := email.New(cfg.ResendAPIKey, cfg.EmailFrom, cfg.EmailFromName, cfg.APIBaseURL)

	authSvc := auth.NewService(pool, tokenManager, emailClient, cfg.MagicLinkHMACSecret, cfg.MagicLinkExpiry, cfg.APIBaseURL)
	memberRepo := members.NewRepository(pool)
	membersSvc := members.NewService(memberRepo)
	xpSvc := xp.NewXPService(pool)
	eventsSvc := events.NewEventsService(pool)
	duesSvc := dues.NewDuesService(pool, cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	notifSvc := notifications.NewNotificationsService(pool)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := auth.NewHandler(authSvc)
	membersHandler := members.NewHandler(membersSvc)
	xpHandler := xp.NewHandler(xpSvc)
	eventsHandler := events.NewHandler(eventsSvc, cfg.MagicLinkHMACSecret) // reusing HMAC secret for QR
	duesHandler := dues.NewHandler(duesSvc, cfg.StripeWebhookSecret)
	notifHandler := notifications.NewHandler(notifSvc)

	// ── Echo Setup ────────────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.New()

	// ── Middleware ────────────────────────────────────────────────────────────
	e.Use(middleware.RequestID())
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			log.Error().Err(err).Str("stack", string(stack)).Msg("panic recovered")
			return nil
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     strings.Split(cfg.AllowedOrigins, ","),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization, "X-Request-ID"},
		AllowCredentials: false,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","method":"${method}","uri":"${uri}","status":${status},"latency_ms":${latency_ms}}` + "\n",
	}))

	jwtMW := auth.JWTMiddleware(tokenManager)

	// ── Routes ────────────────────────────────────────────────────────────────
	v1 := e.Group("/v1")

	// Health check (no auth)
	v1.GET("/healthz", func(c echo.Context) error {
		ctx := c.Request().Context()
		if err := appdb.Ping(ctx, pool); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unhealthy", "error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"})
	})

	// Auth routes
	authGroup := v1.Group("/auth")
	authHandler.RegisterRoutes(authGroup, jwtMW)

	// Webhooks (no JWT)
	webhookGroup := v1.Group("/webhooks")
	duesHandler.RegisterRoutes(v1.Group("/dues"), webhookGroup, jwtMW)

	// Protected routes
	membersHandler.RegisterRoutes(v1.Group("/members"), jwtMW)
	xpHandler.RegisterRoutes(v1, jwtMW)
	eventsHandler.RegisterRoutes(v1.Group("/events"), jwtMW)
	notifHandler.RegisterRoutes(v1.Group("/notifications"), jwtMW)

	// ── Error handler ─────────────────────────────────────────────────────────
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)
		if !ok {
			he = &echo.HTTPError{Code: http.StatusInternalServerError, Message: "internal server error"}
		}

		if !c.Response().Committed {
			c.JSON(he.Code, map[string]interface{}{
				"error": map[string]interface{}{
					"code":    http.StatusText(he.Code),
					"message": fmt.Sprintf("%v", he.Message),
					"status":  he.Code,
				},
			})
		}
	}

	// ── Start server with graceful shutdown ───────────────────────────────────
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Info().Str("addr", addr).Msg("starting server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("server stopped")
}

// runMigrations applies pending database migrations using golang-migrate.
func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
