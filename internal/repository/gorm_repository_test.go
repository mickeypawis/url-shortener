package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mickeypawis/url-shortener/internal/model"
	"github.com/mickeypawis/url-shortener/internal/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("url_shortener_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.URL{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestGormURLRepositoryCreateAndFind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := repository.NewGormURLRepository(db)
	ctx := context.Background()

	u := &model.URL{Code: "abc1234", LongURL: "https://example.com"}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByCode(ctx, "abc1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.LongURL != "https://example.com" {
		t.Fatalf("unexpected long url: %s", found.LongURL)
	}
}

func TestGormURLRepositoryCreateDuplicateCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := repository.NewGormURLRepository(db)
	ctx := context.Background()

	u1 := &model.URL{Code: "dup0001", LongURL: "https://example.com/1"}
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u2 := &model.URL{Code: "dup0001", LongURL: "https://example.com/2"}
	err := repo.Create(ctx, u2)
	if !errors.Is(err, repository.ErrCodeExists) {
		t.Fatalf("expected ErrCodeExists, got %v", err)
	}
}

func TestGormURLRepositoryFindByCodeNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := repository.NewGormURLRepository(db)

	_, err := repo.FindByCode(context.Background(), "missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
