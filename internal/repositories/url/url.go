package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mickeypawis/url-shortener/internal/model"
	"gorm.io/gorm"
)

var (
	ErrNotFound   = errors.New("url not found")
	ErrCodeExists = errors.New("code already exists")
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	FindByCode(ctx context.Context, code string) (*model.URL, error)
	FindAll(ctx context.Context) ([]model.URL, error)
	Delete(ctx context.Context, id uint) error
}

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

func (r *gormURLRepository) FindAll(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	if err := r.db.WithContext(ctx).Find(&urls).Error; err != nil {
		return nil, err
	}
	return urls, nil
}

func (r *gormURLRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.URL{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
