package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/config"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
	"github.com/ojooji/booking-service-api/pkg/hash"
	"github.com/ojooji/booking-service-api/pkg/jwt"
)

var (
	ErrEmailTaken    = errors.New("email already taken")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrUserNotFound  = errors.New("user not found")
	ErrNotAuthorized = errors.New("not authorized")
)

type AuthService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	pwHash, err := hash.Password(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: pwHash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil, ErrInvalidCreds
	}
	if err != nil {
		return "", nil, err
	}

	if !hash.Check(password, user.PasswordHash) {
		return "", nil, ErrInvalidCreds
	}

	token, err := jwt.Generate(user.ID, string(user.Role), s.cfg.JWTSecret, s.cfg.JWTExpiration)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}
