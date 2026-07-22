package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *domain.Employee) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error)
	List(ctx context.Context) ([]domain.Employee, error)
	Update(ctx context.Context, emp *domain.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetServices(ctx context.Context, employeeID uuid.UUID, serviceIDs []uuid.UUID) error
	GetServices(ctx context.Context, employeeID uuid.UUID) ([]uuid.UUID, error)
}
