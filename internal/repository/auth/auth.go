package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	modelauth "github.com/mickeypawis/url-shortener/internal/model"
)

const pgUniqueViolationCode = "23505"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *modelauth.User) error
	FindByEmail(ctx context.Context, email string) (*modelauth.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(ctx context.Context, user *modelauth.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*modelauth.User, error) {
	var user modelauth.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
