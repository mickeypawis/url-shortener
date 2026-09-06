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
	"github.com/mickeypawis/url-shortener/internal/model"
	"github.com/mickeypawis/url-shortener/internal/repository"
	repoauth "github.com/mickeypawis/url-shortener/internal/repository/auth"
	"github.com/mickeypawis/url-shortener/internal/service"
	svcauth "github.com/mickeypawis/url-shortener/internal/service/auth"
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
	handler := api.NewHandler(svc, cfg.BaseURL)

	userRepo := repoauth.NewGormUserRepository(db)
	authSvc := svcauth.NewService(userRepo, cfg.JWTSecret)
	authHandler := apiauth.NewHandler(authSvc)

	router := api.NewRouter(handler, authHandler)

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
