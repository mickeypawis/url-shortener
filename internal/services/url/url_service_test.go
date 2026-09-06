package service_url_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mickeypawis/url-shortener/internal/model"
	repository "github.com/mickeypawis/url-shortener/internal/repositories/url"

	service "github.com/mickeypawis/url-shortener/internal/services/url"
)

type fakeRepo struct {
	byCode map[string]*model.URL
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byCode: make(map[string]*model.URL)}
}

func (f *fakeRepo) Create(_ context.Context, u *model.URL) error {
	if _, exists := f.byCode[u.Code]; exists {
		return repository.ErrCodeExists
	}
	f.byCode[u.Code] = u
	return nil
}

func (f *fakeRepo) FindByCode(_ context.Context, code string) (*model.URL, error) {
	u, ok := f.byCode[code]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func TestShortenValidURL(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())

	u, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Code == "" {
		t.Fatal("expected non-empty code")
	}
	if u.LongURL != "https://example.com" {
		t.Fatalf("unexpected long url: %s", u.LongURL)
	}
}

func TestShortenInvalidURL(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())

	_, err := svc.Shorten(context.Background(), "not-a-url")
	if !errors.Is(err, service.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestResolveNotFound(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())

	_, err := svc.Resolve(context.Background(), "missing")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveFound(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())

	created, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := svc.Resolve(context.Background(), created.Code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.LongURL != "https://example.com" {
		t.Fatalf("unexpected long url: %s", found.LongURL)
	}
}
