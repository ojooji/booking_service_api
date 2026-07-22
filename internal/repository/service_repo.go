package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
)

type ServiceRepository interface {
	Create(ctx context.Context, svc *domain.Service) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error)
	List(ctx context.Context) ([]domain.Service, error)
	Update(ctx context.Context, svc *domain.Service) error
	Delete(ctx context.Context, id uuid.UUID) error
}
