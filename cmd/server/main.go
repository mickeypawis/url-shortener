package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mickeypawis/url-shortener/configs"
	"github.com/mickeypawis/url-shortener/internal/api"
	apiauth "github.com/mickeypawis/url-shortener/internal/api/auth"
	apiurl "github.com/mickeypawis/url-shortener/internal/api/url"
	"github.com/mickeypawis/url-shortener/internal/model"
	repoauth "github.com/mickeypawis/url-shortener/internal/repositories/auth"
	repository "github.com/mickeypawis/url-shortener/internal/repositories/url"
	svcauth "github.com/mickeypawis/url-shortener/internal/services/auth"
	service "github.com/mickeypawis/url-shortener/internal/services/url"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := configs.Load()

	if cfg.JWTSecret == "" {
		slog.Error("JWT_SECRET must be set")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.URL{}, &model.User{}); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	repo := repository.NewGormURLRepository(db)
	svc := service.NewURLService(repo)
	handler := apiurl.NewHandler(svc, cfg.BaseURL)

	userRepo := repoauth.NewGormUserRepository(db)
	authSvc := svcauth.NewService(userRepo, cfg.JWTSecret)
	authHandler := apiauth.NewHandler(authSvc)
	authMiddleware := apiauth.Middleware(authSvc)

	router := api.NewRouter(handler, authHandler, authMiddleware)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		slog.Info("starting server", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
}
