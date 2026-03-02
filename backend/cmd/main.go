package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"media-cms/internal/config"
	"media-cms/internal/handler"
	"media-cms/internal/middleware"
	"media-cms/internal/repository"
	"media-cms/internal/service"
	"media-cms/internal/storage"
)

func main() {
	// ── 1. Load .env ────────────────────────────────────────────────────────
	_ = godotenv.Load()

	// ── 2. Config ───────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// ── 3. Logger ───────────────────────────────────────────────────────────
	log := buildLogger(cfg.Server.Mode)
	defer log.Sync() //nolint:errcheck

	// ── 4. Database ─────────────────────────────────────────────────────────
	db, err := sqlx.Open("postgres", cfg.Database.DSN)
	if err != nil {
		log.Fatal("db open failed", zap.Error(err))
	}
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.MaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatal("db ping failed", zap.Error(err))
	}
	log.Info("connected to PostgreSQL")

	// ── 5. MinIO ────────────────────────────────────────────────────────────
	minioStore, err := storage.NewMinIOStorage(cfg.MinIO)
	if err != nil {
		log.Fatal("minio init failed", zap.Error(err))
	}
	// Ensure default bucket exists at startup
	if err = minioStore.EnsureBucket(context.Background(), cfg.MinIO.DefaultBucket); err != nil {
		log.Fatal("minio ensure bucket failed", zap.Error(err))
	}
	log.Info("connected to MinIO")

	// ── 6. Repositories ─────────────────────────────────────────────────────
	mediaRepo := repository.NewMediaRepository(db)
	refRepo := repository.NewReferenceRepository(db)
	authRepo := repository.NewAuthRepository(db)
	if err = authRepo.EnsureProjectPolicySchema(context.Background()); err != nil {
		log.Fatal("ensure project policy schema failed", zap.Error(err))
	}
	if err = authRepo.EnsureUploadLogSchema(context.Background()); err != nil {
		log.Fatal("ensure upload log schema failed", zap.Error(err))
	}

	// ── 8. Services ─────────────────────────────────────────────────────────
	mediaSvc := service.NewMediaService(mediaRepo, authRepo, minioStore, cfg, log)
	refSvc := service.NewReferenceService(refRepo, mediaRepo, log)
	authSvc := service.NewAuthService(authRepo, cfg, log)
	if err = authSvc.EnsureDefaultAdmin(context.Background()); err != nil {
		log.Fatal("ensure default admin failed", zap.Error(err))
	}

	// ── 9. Handlers ─────────────────────────────────────────────────────────
	mediaHandler := handler.NewMediaHandler(mediaSvc, log)
	refHandler := handler.NewReferenceHandler(refSvc, log)
	authHandler := handler.NewAuthHandler(authSvc, log)

	// ── 10. Router ──────────────────────────────────────────────────────────
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logger(log))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	r.GET("/health", handler.HealthCheck)
	r.GET("/docs/openapi.yaml", handler.ServeOpenAPISpec)
	r.GET("/docs/swagger", handler.ServeSwaggerUI)
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/project-token", authHandler.IssueProjectToken)
	}

	// Admin routes
	admin := r.Group("/auth")
	admin.Use(middleware.JWT(cfg.JWT.Secret), middleware.RequireRole("admin"))
	{
		admin.GET("/users", authHandler.ListUsers)
		admin.POST("/users", authHandler.CreateUser)
		admin.PUT("/users/:id", authHandler.UpdateUser)
		admin.DELETE("/users/:id", authHandler.DeleteUser)
		admin.GET("/projects", authHandler.ListProjects)
		admin.POST("/projects", authHandler.CreateProject)
		admin.PUT("/projects/:id", authHandler.UpdateProject)
		admin.DELETE("/projects/:id", authHandler.DeleteProject)
		admin.GET("/projects/:id/upload-logs", authHandler.ListProjectUploadLogs)
	}

	// Routes — JWT is optional, controlled by AUTH_REQUIRED env var
	var apiMiddleware []gin.HandlerFunc
	if cfg.JWT.Required {
		apiMiddleware = append(apiMiddleware, middleware.JWT(cfg.JWT.Secret))
	}

	api := r.Group("/media", apiMiddleware...)
	{
		api.POST("/upload", mediaHandler.Upload)
		api.GET("", mediaHandler.List)
		api.GET("/filter-options", mediaHandler.FilterOptions)
		api.GET("/:id", mediaHandler.GetByID)
		api.DELETE("/:id", mediaHandler.Delete)
		api.GET("/:id/usage", refHandler.GetUsage)

		api.POST("/reference", refHandler.AddReference)
		api.DELETE("/reference", refHandler.RemoveReference)
	}

	// ── 11. Cleanup Cron ────────────────────────────────────────────────────
	c := cron.New()
	_, err = c.AddFunc("@every 1h", func() {
		n, cronErr := mediaSvc.CleanupOrphans(context.Background())
		if cronErr != nil {
			log.Error("cleanup cron error", zap.Error(cronErr))
		} else {
			log.Info("cleanup cron done", zap.Int("deleted", n))
		}
	})
	if err != nil {
		log.Warn("failed to register cleanup cron", zap.Error(err))
	}
	c.Start()

	// ── 12. HTTP Server ─────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr))
		if serveErr := srv.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(serveErr))
		}
	}()

	// ── 13. Graceful Shutdown ────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	c.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped")
}

func buildLogger(mode string) *zap.Logger {
	var cfg zap.Config
	if mode == "release" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	log, _ := cfg.Build()
	return log
}
