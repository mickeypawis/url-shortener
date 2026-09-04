package repository

import (
	"context"
	"errors"

	"github.com/mickeypawis/url-shortener/internal/model"
)

var (
	ErrNotFound   = errors.New("url not found")
	ErrCodeExists = errors.New("code already exists")
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	FindByCode(ctx context.Context, code string) (*model.URL, error)
}
