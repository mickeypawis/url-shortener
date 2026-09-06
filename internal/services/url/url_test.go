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

func (f *fakeRepo) FindAll(_ context.Context) ([]model.URL, error) {
	urls := make([]model.URL, 0, len(f.byCode))
	for _, u := range f.byCode {
		urls = append(urls, *u)
	}
	return urls, nil
}

func (f *fakeRepo) Delete(_ context.Context, id uint) error {
	for code, u := range f.byCode {
		if u.ID == id {
			delete(f.byCode, code)
			return nil
		}
	}
	return repository.ErrNotFound
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

func TestList(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())
	ctx := context.Background()

	if _, err := svc.Shorten(ctx, "https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Shorten(ctx, "https://example.org"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	urls, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())

	err := svc.Delete(context.Background(), 999)
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteFound(t *testing.T) {
	svc := service.NewURLService(newFakeRepo())
	ctx := context.Background()

	created, err := svc.Shorten(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	created.ID = 1

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := svc.Resolve(ctx, created.Code); !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected url to be deleted, got err=%v", err)
	}
}
