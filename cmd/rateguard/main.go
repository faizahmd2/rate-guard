package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/faizahmd2/rate-guard/internal/api/handler"
	"github.com/faizahmd2/rate-guard/internal/api/middleware"
	"github.com/faizahmd2/rate-guard/internal/auth"
	"github.com/faizahmd2/rate-guard/internal/config"
	"github.com/faizahmd2/rate-guard/internal/limiter"
	"github.com/faizahmd2/rate-guard/internal/logger"
	"github.com/faizahmd2/rate-guard/internal/postgres"
	"github.com/faizahmd2/rate-guard/internal/redis"
	"github.com/faizahmd2/rate-guard/internal/sqlite"
	"github.com/faizahmd2/rate-guard/internal/web"
)

func main() {
	appLogger := logger.New()

	cfg, err := config.Load()
	if err != nil {
		appLogger.Error(
			"failed to load configuration",
			"error", err,
		)

		os.Exit(1)
	}

	ctx := context.Background()

	redis, err := redis.NewClient(
		ctx,
		cfg.Redis.Host,
		cfg.Redis.Port,
	)
	if err != nil {
		appLogger.Error(
			"failed to connect to redis",
			"error", err,
		)

		os.Exit(1)
	}
	defer redis.Close()

	var ruleRepository config.RuleRepository
	var sqlDB *sql.DB

	// Admin authentication currently uses SQLite.
	var authHandler *handler.AuthHandler
	var authService *auth.Service
	var cookieManager *auth.CookieManager
	var tokenStore *auth.TokenStore
	var systemHandler *handler.SystemHandler

	switch cfg.StorageDriver() {
	case "sqlite":
		sqlDB, err = sqlite.NewDB()
		if err != nil {
			appLogger.Error(
				"failed to connect to sqlite",
				"error", err,
			)

			os.Exit(1)
		}

		defer sqlDB.Close()

		ruleRepository = config.NewSQLiteRuleRepository(
			sqlDB,
		)

		authRepository := auth.NewSQLiteRepository(sqlDB)

		authService = auth.NewService(
			authRepository,
		)

		cookieManager = auth.NewCookieManager(
			cfg.Auth.CookieSecret,
		)

		authHandler = handler.NewAuthHandler(
			authService,
			cookieManager,
		)

		activeTokens, err := authRepository.GetActiveTokens(ctx)
		if err != nil {
			appLogger.Error("failed to load api tokens", "error", err)
			return
		}

		tokenStore = auth.NewTokenStore(activeTokens)

	case "postgres":
		db, err := postgres.NewPool(
			ctx,
			cfg.Storage.Postgres,
		)
		if err != nil {
			appLogger.Error(
				"failed to connect to postgres",
				"error", err,
			)

			os.Exit(1)
		}

		defer db.Close()

		ruleRepository = config.NewPostgresRuleRepository(
			db,
		)

	default:
		appLogger.Error(
			"unsupported storage driver",
			"driver", cfg.StorageDriver(),
		)

		os.Exit(1)
	}

	if authService == nil || cookieManager == nil || tokenStore == nil {
		appLogger.Error("authentication is not initialized")
		return
	}

	systemHandler = handler.NewSystemHandler(
		cfg.AppEnv,
		cfg.StorageDriver(),
		cfg.Redis.Host,
		cfg.Redis.Port,
		redis,
	)

	ruleCache := config.NewRuleCache(
		redis,
		ruleRepository,
		5*time.Minute,
		30*time.Minute,
	)

	ruleService := config.NewRuleService(
		ruleRepository,
		ruleCache,
	)

	ruleHandler := handler.NewRuleHandler(
		ruleService,
	)

	tokenHandler := handler.NewTokenHandler(
		authService,
		tokenStore,
	)

	limiterService := limiter.NewService(
		ruleCache,
		redis,
	)

	checkHandler := handler.NewCheckHandler(
		limiterService,
	)

	webHandler, err := web.Handler()
	if err != nil {
		appLogger.Error(
			"failed to initialize web frontend",
			"error", err,
		)

		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger(appLogger))
	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.RealIP)
	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.Timeout(10 * time.Second))

	router.Get(
		"/health",
		handler.Health,
	)

	router.Group(func(r chi.Router) {
		r.Use(middleware.APIToken(tokenStore))

		r.Post("/v1/check", checkHandler.Check)
	})

	// Admin authentication.
	//
	// Currently admin auth is available when SQLite is being used.
	if authHandler != nil && cookieManager != nil {
		router.Get(
			"/api/auth/status",
			authHandler.Status,
		)

		router.Post(
			"/api/auth/setup",
			authHandler.Setup,
		)

		router.Post(
			"/api/auth/login",
			authHandler.Login,
		)

		router.Post(
			"/api/auth/logout",
			authHandler.Logout,
		)

		router.Group(func(r chi.Router) {
			r.Use(
				middleware.AdminAuth(cookieManager),
			)

			r.Get("/api/system", systemHandler.Get)
			r.Post("/v1/rules", ruleHandler.Create)
			r.Get("/v1/rules", ruleHandler.List)
			r.Get("/v1/rules/{service}/{resource}", ruleHandler.Get)
			r.Put("/v1/rules/{service}/{resource}", ruleHandler.Update)
			r.Delete("/v1/rules/{service}/{resource}", ruleHandler.Delete)
		})

		router.Group(func(r chi.Router) {
			r.Use(middleware.AdminAuth(cookieManager))

			r.Post("/api/tokens", tokenHandler.Create)
			r.Get("/api/tokens", tokenHandler.List)
			r.Delete("/api/tokens/{id}", tokenHandler.Revoke)
		})
	}

	router.NotFound(webHandler.ServeHTTP)

	addr := ":" + cfg.Port

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		appLogger.Info(
			"RateGuard starting",
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			appLogger.Error(
				"server failed",
				"error", err,
			)

			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	appLogger.Info(
		"shutdown signal received",
	)

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error(
			"server shutdown failed",
			"error", err,
		)
	}

	appLogger.Info(
		"RateGuard stopped",
	)
}
