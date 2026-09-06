package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	modelauth "github.com/mickeypawis/url-shortener/internal/model"
	repoauth "github.com/mickeypawis/url-shortener/internal/repository/auth"
)

var (
	ErrEmailExists        = repoauth.ErrEmailExists
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const tokenTTL = 24 * time.Hour

type Service struct {
	repo      repoauth.UserRepository
	jwtSecret []byte
}

func NewService(repo repoauth.UserRepository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (s *Service) Register(ctx context.Context, email, password string) (*modelauth.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &modelauth.User{Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repoauth.ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repoauth.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.issueToken(user.ID)
}

func (s *Service) issueToken(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
