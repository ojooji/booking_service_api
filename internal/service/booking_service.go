package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/dto"
	"github.com/ojooji/booking-service-api/internal/repository"
)

var (
	ErrConflict           = errors.New("time slot conflict: employee already booked")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrPastStartTime      = errors.New("start time must be in the future")
	ErrOutsideHours       = errors.New("booking must be within business hours (09:00-18:00)")
	ErrUnalignedStartTime = errors.New("start time must align to 30-minute slots")
)

// Business hours and slot grid shared by slot listing and booking validation.
const (
	businessOpenHour  = 9
	businessCloseHour = 18
	slotInterval      = 30 * time.Minute
)

type BookingService struct {
	bookingRepo  repository.BookingRepository
	employeeRepo repository.EmployeeRepository
	serviceRepo  repository.ServiceRepository
}

func NewBookingService(bookingRepo repository.BookingRepository, employeeRepo repository.EmployeeRepository, serviceRepo repository.ServiceRepository) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		employeeRepo: employeeRepo,
		serviceRepo:  serviceRepo,
	}
}

func (s *BookingService) Create(ctx context.Context, userID, employeeID, serviceID uuid.UUID, startTime time.Time) (*domain.Booking, error) {
	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	if _, err := s.employeeRepo.GetByID(ctx, employeeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	endTime := startTime.Add(time.Duration(svc.Duration) * time.Minute)

	if err := validateStartTime(startTime, endTime); err != nil {
		return nil, err
	}

	now := time.Now()
	booking := &domain.Booking{
		ID:         uuid.New(),
		UserID:     userID,
		EmployeeID: employeeID,
		ServiceID:  serviceID,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     domain.StatusConfirmed,
		TotalPrice: svc.Price,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.bookingRepo.CreateAtomic(ctx, booking, employeeID, startTime, endTime); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrConflict
		}
		return nil, err
	}

	return booking, nil
}

func validateStartTime(startTime, endTime time.Time) error {
	if !startTime.After(time.Now()) {
		return ErrPastStartTime
	}
	if startTime.Minute()%int(slotInterval.Minutes()) != 0 || startTime.Second() != 0 || startTime.Nanosecond() != 0 {
		return ErrUnalignedStartTime
	}
	dayStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	open := dayStart.Add(businessOpenHour * time.Hour)
	close := dayStart.Add(businessCloseHour * time.Hour)
	if startTime.Before(open) || endTime.After(close) {
		return ErrOutsideHours
	}
	return nil
}

// GetByID returns a booking, restricted to its owner or an admin.
func (s *BookingService) GetByID(ctx context.Context, id, requesterID uuid.UUID, role string) (*domain.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	if b.UserID != requesterID && role != string(domain.RoleAdmin) {
		return nil, ErrNotAuthorized
	}
	return b, nil
}

func (s *BookingService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	return s.bookingRepo.ListByUser(ctx, userID)
}

func (s *BookingService) ListAll(ctx context.Context) ([]domain.Booking, error) {
	return s.bookingRepo.ListAll(ctx)
}

func (s *BookingService) Cancel(ctx context.Context, id uuid.UUID, userID uuid.UUID, role string) error {
	b, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		return err
	}
	if b.UserID != userID && role != string(domain.RoleAdmin) {
		return ErrNotAuthorized
	}
	return s.bookingRepo.UpdateStatus(ctx, id, domain.StatusCancelled)
}

func (s *BookingService) GetSlots(ctx context.Context, employeeID uuid.UUID, date time.Time) ([]dto.TimeSlot, error) {
	if _, err := s.employeeRepo.GetByID(ctx, employeeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	dayBookings, err := s.bookingRepo.ListByEmployeeAndDate(ctx, employeeID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	slots := []dto.TimeSlot{}
	slotStart := dayStart.Add(businessOpenHour * time.Hour)
	slotEnd := dayStart.Add(businessCloseHour * time.Hour)

	current := slotStart
	for current.Before(slotEnd) {
		slotEndTime := current.Add(slotInterval)
		if slotEndTime.After(slotEnd) {
			break
		}
		conflict := false
		for _, b := range dayBookings {
			if current.Before(b.EndTime) && slotEndTime.After(b.StartTime) {
				conflict = true
				if b.EndTime.After(current) {
					current = b.EndTime
				} else {
					current = slotEndTime
				}
				break
			}
		}
		if !conflict {
			slots = append(slots, dto.TimeSlot{Start: current, End: slotEndTime})
			current = slotEndTime
		}
	}

	return slots, nil
}
