package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
)

var ErrConflict = errors.New("time slot conflict: employee already booked")

type BookingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error)
	ListByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]domain.Booking, error)
	ListAll(ctx context.Context) ([]domain.Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
	CreateAtomic(ctx context.Context, booking *domain.Booking, employeeID uuid.UUID, start, end time.Time) error
}
