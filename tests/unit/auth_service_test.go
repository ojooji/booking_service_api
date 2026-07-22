package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ojooji/booking-service-api/internal/config"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
	"github.com/ojooji/booking-service-api/internal/service"
	"github.com/ojooji/booking-service-api/pkg/hash"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAuthRegister_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	cfg := &config.Config{JWTSecret: "test", JWTExpiration: 3600}
	svc := service.NewAuthService(mockRepo, cfg)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, repository.ErrNotFound)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "test@example.com"
	})).Return(nil)

	user, err := svc.Register(context.Background(), "test@example.com", "password123", "John", "Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, domain.RoleUser, user.Role)
	mockRepo.AssertExpectations(t)
}

func TestAuthRegister_EmailTaken(t *testing.T) {
	mockRepo := new(mockUserRepo)
	cfg := &config.Config{JWTSecret: "test", JWTExpiration: 3600}
	svc := service.NewAuthService(mockRepo, cfg)

	existing := &domain.User{Email: "existing@example.com"}
	mockRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existing, nil)

	user, err := svc.Register(context.Background(), "existing@example.com", "password123", "Jane", "Doe")
	assert.Error(t, err)
	assert.Equal(t, service.ErrEmailTaken, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthLogin_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiration: 3600}
	svc := service.NewAuthService(mockRepo, cfg)

	hashed, _ := hash.Password("password123")
	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashed,
		Role:         domain.RoleUser,
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, returned, err := svc.Login(context.Background(), "test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, user.Email, returned.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthLogin_UnknownEmail(t *testing.T) {
	mockRepo := new(mockUserRepo)
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiration: 3600}
	svc := service.NewAuthService(mockRepo, cfg)

	mockRepo.On("GetByEmail", mock.Anything, "nobody@example.com").Return(nil, repository.ErrNotFound)

	token, returned, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	assert.Error(t, err)
	assert.Equal(t, service.ErrInvalidCreds, err)
	assert.Empty(t, token)
	assert.Nil(t, returned)
	mockRepo.AssertExpectations(t)
}

func TestAuthLogin_InvalidPassword(t *testing.T) {
	mockRepo := new(mockUserRepo)
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiration: 3600}
	svc := service.NewAuthService(mockRepo, cfg)

	hashed, _ := hash.Password("correctpassword")
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: hashed,
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, returned, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, service.ErrInvalidCreds, err)
	assert.Empty(t, token)
	assert.Nil(t, returned)
	mockRepo.AssertExpectations(t)
}
