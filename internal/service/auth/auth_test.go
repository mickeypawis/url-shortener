package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	modelauth "github.com/mickeypawis/url-shortener/internal/model"
	repoauth "github.com/mickeypawis/url-shortener/internal/repository/auth"
	"github.com/mickeypawis/url-shortener/internal/service/auth"
)

type fakeUserRepo struct {
	byEmail map[string]*modelauth.User
	nextID  uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: make(map[string]*modelauth.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, u *modelauth.User) error {
	if _, exists := f.byEmail[u.Email]; exists {
		return repoauth.ErrEmailExists
	}
	f.nextID++
	u.ID = f.nextID
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*modelauth.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, repoauth.ErrUserNotFound
	}
	return u, nil
}

func TestRegisterSuccess(t *testing.T) {
	svc := auth.NewService(newFakeUserRepo(), "test-secret")

	u, err := svc.Register(context.Background(), "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("unexpected email: %s", u.Email)
	}
	if u.PasswordHash == "" || u.PasswordHash == "password123" {
		t.Fatalf("expected password to be hashed, got %q", u.PasswordHash)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc := auth.NewService(newFakeUserRepo(), "test-secret")
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Register(ctx, "alice@example.com", "password456")
	if !errors.Is(err, auth.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	svc := auth.NewService(newFakeUserRepo(), "test-secret")
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokenStr, err := svc.Login(ctx, "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	parsed, err := jwt.Parse(tokenStr, func(*jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("expected valid token, err=%v", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc := auth.NewService(newFakeUserRepo(), "test-secret")
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Login(ctx, "alice@example.com", "wrong-password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	svc := auth.NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Login(context.Background(), "missing@example.com", "password123")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
