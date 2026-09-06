package auth_test

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

	modelauth "github.com/mickeypawis/url-shortener/internal/model"
	"github.com/mickeypawis/url-shortener/internal/repository/auth"
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

	if err := db.AutoMigrate(&modelauth.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestGormUserRepositoryCreateAndFind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := auth.NewGormUserRepository(db)
	ctx := context.Background()

	u := &modelauth.User{Email: "alice@example.com", PasswordHash: "hashed"}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.PasswordHash != "hashed" {
		t.Fatalf("unexpected password hash: %s", found.PasswordHash)
	}
}

func TestGormUserRepositoryCreateDuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := auth.NewGormUserRepository(db)
	ctx := context.Background()

	u1 := &modelauth.User{Email: "alice@example.com", PasswordHash: "hashed1"}
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u2 := &modelauth.User{Email: "alice@example.com", PasswordHash: "hashed2"}
	err := repo.Create(ctx, u2)
	if !errors.Is(err, auth.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestGormUserRepositoryFindByEmailNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := auth.NewGormUserRepository(db)

	_, err := repo.FindByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
