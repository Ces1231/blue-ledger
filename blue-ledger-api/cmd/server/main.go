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

	"github.com/ces1231/blue-ledger-api/internal/ai"
	"github.com/ces1231/blue-ledger-api/internal/alumni"
	"github.com/ces1231/blue-ledger-api/internal/challenges"
	"github.com/ces1231/blue-ledger-api/internal/announcements"
	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/ces1231/blue-ledger-api/internal/badges"
	"github.com/ces1231/blue-ledger-api/internal/committees"
	"github.com/ces1231/blue-ledger-api/internal/dues"
	"github.com/ces1231/blue-ledger-api/internal/fundraising"
	"github.com/ces1231/blue-ledger-api/internal/jobboard"
	"github.com/ces1231/blue-ledger-api/internal/events"
	"github.com/ces1231/blue-ledger-api/internal/goals"
	"github.com/ces1231/blue-ledger-api/internal/health"
	"github.com/ces1231/blue-ledger-api/internal/intake"
	"github.com/ces1231/blue-ledger-api/internal/members"
	"github.com/ces1231/blue-ledger-api/internal/mentorship"
	"github.com/ces1231/blue-ledger-api/internal/milestones"
	"github.com/ces1231/blue-ledger-api/internal/messages"
	"github.com/ces1231/blue-ledger-api/internal/minutes"
	"github.com/ces1231/blue-ledger-api/internal/notifications"
	"github.com/ces1231/blue-ledger-api/internal/platform"
	"github.com/ces1231/blue-ledger-api/internal/props"
	"github.com/ces1231/blue-ledger-api/internal/scholarships"
	"github.com/ces1231/blue-ledger-api/internal/servicelog"
	"github.com/ces1231/blue-ledger-api/internal/settings"
	"github.com/ces1231/blue-ledger-api/internal/store"
	"github.com/ces1231/blue-ledger-api/internal/streaks"
	"github.com/ces1231/blue-ledger-api/internal/quests"
	"github.com/ces1231/blue-ledger-api/internal/resources"
	"github.com/ces1231/blue-ledger-api/internal/sbc"
	"github.com/ces1231/blue-ledger-api/internal/studygroups"
	"github.com/ces1231/blue-ledger-api/internal/votes"
	"github.com/ces1231/blue-ledger-api/internal/xp"
	"github.com/ces1231/blue-ledger-api/pkg/config"
	appdb "github.com/ces1231/blue-ledger-api/pkg/db"
	"github.com/ces1231/blue-ledger-api/pkg/email"
	"github.com/ces1231/blue-ledger-api/pkg/logger"
	appmiddleware "github.com/ces1231/blue-ledger-api/pkg/middleware"
	appredis "github.com/ces1231/blue-ledger-api/pkg/redis"
	"github.com/ces1231/blue-ledger-api/pkg/validator"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	rateLimitMW := appmiddleware.RateLimiter(redisClient)

	// ── Run Migrations ────────────────────────────────────────────────────────
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("failed to run database migrations")
	}
	log.Info().Msg("migrations applied")

	// ── WebSocket Hub ────────────────────────────────────────────────────────
	var wsHub *platform.Hub
	if redisClient != nil {
		wsHub = platform.NewHub(redisClient)
		go wsHub.Run()
		log.Info().Msg("websocket hub started")
	}

	// broadcastFn wires the hub into domain services (nil-safe)
	broadcastFn := func(chapterID, memberID, msgType string, payload any) {
		if wsHub == nil {
			return
		}
		if memberID != "" {
			wsHub.SendToMember(memberID, msgType, payload)
		} else {
			wsHub.BroadcastToChapter(chapterID, msgType, payload)
		}
	}

	// ── Core services ─────────────────────────────────────────────────────────
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
	streaksSvc := streaks.NewService(pool)
	xpSvc := xp.NewXPService(pool, streaksSvc)
	eventsSvc := events.NewEventsService(pool)
	duesSvc := dues.NewDuesService(pool, cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	notifSvc := notifications.NewNotificationsService(pool)

	// ── New domain services ───────────────────────────────────────────────────
	badgesSvc := badges.NewService(pool)
	questsSvc := quests.NewService(pool)
	committeesSvc := committees.NewService(pool)
	fundraisingSvc := fundraising.NewService(pool)
	resourcesSvc := resources.NewService(pool)
	jobboardSvc := jobboard.NewService(pool)
	alumniSvc := alumni.NewService(pool)
	sbcSvc := sbc.NewService(pool)
	challengesSvc := challenges.NewService(pool, xpSvc, broadcastFn)
	aiSvc := ai.NewService(pool, cfg.AnthropicAPIKey, cfg.OpenAIAPIKey)
	announcementsSvc := announcements.NewService(pool)
	propsSvc := props.NewService(pool, xpSvc)
	servicelogSvc := servicelog.NewService(pool, xpSvc)
	intakeSvc := intake.NewService(pool)
	votesSvc := votes.NewService(pool)
	mentorshipSvc := mentorship.NewService(pool, xpSvc)
	minutesSvc := minutes.NewService(pool)
	scholarshipsSvc := scholarships.NewService(pool)
	storeSvc := store.NewService(pool)
	goalsSvc := goals.NewService(pool)
	messagesSvc := messages.NewService(pool)
	milestonesSvc := milestones.NewService(pool)
	studygroupsSvc := studygroups.NewService(pool)
	platformSvc := platform.NewSysService(pool)
	healthSvc := health.NewService(pool)
	settingsSvc := settings.NewService(pool)

	// ── Repositories ──────────────────────────────────────────────────────────
	loginBonusRepo := auth.NewLoginBonusRepository(pool)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := auth.NewHandler(authSvc, loginBonusRepo)
	membersHandler := members.NewHandler(membersSvc)
	xpHandler := xp.NewHandler(xpSvc)
	eventsHandler := events.NewHandler(eventsSvc, cfg.MagicLinkHMACSecret, streaksSvc)
	streaksHandler := streaks.NewHandler(streaksSvc)
	duesHandler := dues.NewHandler(duesSvc, cfg.StripeWebhookSecret)
	notifHandler := notifications.NewHandler(notifSvc)

	badgesHandler := badges.NewHandler(badgesSvc)
	questsHandler := quests.NewHandler(questsSvc)
	committeesHandler := committees.NewHandler(committeesSvc)
	fundraisingHandler := fundraising.NewHandler(fundraisingSvc)
	resourcesHandler := resources.NewHandler(resourcesSvc)
	jobboardHandler := jobboard.NewHandler(jobboardSvc)
	alumniHandler := alumni.NewHandler(alumniSvc)
	sbcHandler := sbc.NewHandler(sbcSvc)
	challengesHandler := challenges.NewHandler(challengesSvc)
	aiHandler := ai.NewHandler(aiSvc, cfg.AnthropicAPIKey, cfg.OpenAIAPIKey)
	announcementsHandler := announcements.NewHandler(announcementsSvc)
	propsHandler := props.NewHandler(propsSvc)
	servicelogHandler := servicelog.NewHandler(servicelogSvc)
	intakeHandler := intake.NewHandler(intakeSvc)
	votesHandler := votes.NewHandler(votesSvc)
	mentorshipHandler := mentorship.NewHandler(mentorshipSvc)
	minutesHandler := minutes.NewHandler(minutesSvc)
	scholarshipsHandler := scholarships.NewHandler(scholarshipsSvc)
	storeHandler := store.NewHandler(storeSvc)
	goalsHandler := goals.NewHandler(goalsSvc)
	messagesHandler := messages.NewHandler(messagesSvc)
	milestonesHandler := milestones.NewHandler(milestonesSvc)
	studygroupsHandler := studygroups.NewHandler(studygroupsSvc)
	sysHandler := platform.NewSysHandler(platformSvc)
	billingHandler := platform.NewBillingHandler(cfg.StripeSecretKey, cfg.APIBaseURL)
	healthHandler := health.NewHandler(healthSvc)
	settingsHandler := settings.NewHandler(settingsSvc)

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
	e.Use(rateLimitMW) // 100 req/min per IP (no-op when Redis is unavailable)

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

	// Core protected routes
	membersHandler.RegisterRoutes(v1.Group("/members"), jwtMW)
	xpHandler.RegisterRoutes(v1, jwtMW)
	eventsHandler.RegisterRoutes(v1.Group("/events"), jwtMW)
	streaksHandler.RegisterRoutes(v1.Group("/streaks"), jwtMW)
	notifHandler.RegisterRoutes(v1.Group("/notifications"), jwtMW)

	// New domain routes
	announcementsHandler.RegisterRoutes(v1.Group("/announcements"), jwtMW)
	propsHandler.RegisterRoutes(v1.Group("/props"), jwtMW)
	servicelogHandler.RegisterRoutes(v1.Group("/service"), jwtMW)
	intakeHandler.RegisterRoutes(v1.Group("/intake"), jwtMW)
	votesHandler.RegisterRoutes(v1.Group("/votes"), jwtMW)
	mentorshipHandler.RegisterRoutes(v1.Group("/mentorship"), jwtMW)
	minutesHandler.RegisterRoutes(v1.Group("/minutes"), jwtMW)
	scholarshipsHandler.RegisterRoutes(v1.Group("/scholarships"), jwtMW)
	storeHandler.RegisterRoutes(v1.Group("/store"), jwtMW)
	goalsHandler.RegisterRoutes(v1.Group("/goals"), jwtMW)
	messagesHandler.RegisterRoutes(v1.Group("/messages"), jwtMW)
	milestonesHandler.RegisterRoutes(v1.Group("/milestones"), jwtMW)
	studygroupsHandler.RegisterRoutes(v1.Group("/study-groups"), jwtMW)
	healthHandler.RegisterRoutes(v1.Group("/health"), jwtMW)
	settingsHandler.RegisterRoutes(v1.Group("/settings"), jwtMW)

	// Gap sprint routes
	badgesHandler.RegisterRoutes(v1.Group("/badges"), jwtMW)
	questsHandler.RegisterRoutes(v1.Group("/quests"), jwtMW)
	committeesHandler.RegisterRoutes(v1.Group("/committees"), jwtMW)
	fundraisingHandler.RegisterRoutes(v1.Group("/fundraising"), jwtMW)
	resourcesHandler.RegisterRoutes(v1.Group("/resources"), jwtMW)
	jobboardHandler.RegisterRoutes(v1.Group("/job-board"), jwtMW)
	alumniHandler.RegisterRoutes(v1.Group("/alumni"), jwtMW)
	sbcHandler.RegisterRoutes(v1.Group("/sbc"), jwtMW)
	challengesHandler.RegisterRoutes(v1.Group("/challenges"), jwtMW)

	// WebSocket + presence routes
	if wsHub != nil {
		v1.GET("/ws", wsHub.HandleWebSocket, jwtMW)
		v1.GET("/presence", wsHub.HandlePresenceList, jwtMW)
	}

	// AI assistant routes
	aiHandler.RegisterRoutes(v1.Group("/ai"), jwtMW)

	// Platform / sysadmin routes
	sysHandler.RegisterSysRoutes(v1, jwtMW)
	billingHandler.RegisterBillingRoutes(v1, jwtMW)

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


