package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/mickeypawis/url-shortener/internal/model"
	"github.com/mickeypawis/url-shortener/internal/repository"
	"github.com/mickeypawis/url-shortener/pkg/base62"
)

var (
	ErrInvalidURL       = errors.New("invalid url")
	ErrNotFound         = repository.ErrNotFound
	errCodeGenExhausted = errors.New("failed to generate a unique code")
)

const (
	codeLength          = 7
	maxGenerateAttempts = 5
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(ctx context.Context, longURL string) (*model.URL, error) {
	if !isValidURL(longURL) {
		return nil, ErrInvalidURL
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		code, err := base62.Generate(codeLength)
		if err != nil {
			return nil, err
		}

		u := &model.URL{Code: code, LongURL: longURL}
		if err := s.repo.Create(ctx, u); err != nil {
			if errors.Is(err, repository.ErrCodeExists) {
				continue
			}
			return nil, err
		}
		return u, nil
	}
	return nil, errCodeGenExhausted
}

func (s *URLService) Resolve(ctx context.Context, code string) (*model.URL, error) {
	return s.repo.FindByCode(ctx, code)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
