package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/mickeypawis/url-shortener/internal/model"
)

const pgUniqueViolationCode = "23505"

type gormURLRepository struct {
	db *gorm.DB
}

func NewGormURLRepository(db *gorm.DB) URLRepository {
	return &gormURLRepository{db: db}
}

func (r *gormURLRepository) Create(ctx context.Context, url *model.URL) error {
	if err := r.db.WithContext(ctx).Create(url).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrCodeExists
		}
		return err
	}
	return nil
}

func (r *gormURLRepository) FindByCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&url).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
