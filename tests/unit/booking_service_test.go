package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
	"github.com/ojooji/booking-service-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBookingRepo struct {
	mock.Mock
}

func (m *mockBookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) ListAll(ctx context.Context) ([]domain.Booking, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockBookingRepo) ListByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]domain.Booking, error) {
	args := m.Called(ctx, employeeID, start, end)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) CreateAtomic(ctx context.Context, booking *domain.Booking, employeeID uuid.UUID, start, end time.Time) error {
	args := m.Called(ctx, booking, employeeID, start, end)
	return args.Error(0)
}

type mockServiceRepo struct {
	mock.Mock
}

func (m *mockServiceRepo) Create(ctx context.Context, svc *domain.Service) error {
	args := m.Called(ctx, svc)
	return args.Error(0)
}

func (m *mockServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Service), args.Error(1)
}

func (m *mockServiceRepo) List(ctx context.Context) ([]domain.Service, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Service), args.Error(1)
}

func (m *mockServiceRepo) Update(ctx context.Context, svc *domain.Service) error {
	args := m.Called(ctx, svc)
	return args.Error(0)
}

func (m *mockServiceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockEmployeeRepo struct {
	mock.Mock
}

func (m *mockEmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

func (m *mockEmployeeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Employee), args.Error(1)
}

func (m *mockEmployeeRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.Employee), args.Error(1)
}

func (m *mockEmployeeRepo) List(ctx context.Context) ([]domain.Employee, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Employee), args.Error(1)
}

func (m *mockEmployeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

func (m *mockEmployeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEmployeeRepo) SetServices(ctx context.Context, employeeID uuid.UUID, serviceIDs []uuid.UUID) error {
	args := m.Called(ctx, employeeID, serviceIDs)
	return args.Error(0)
}

func (m *mockEmployeeRepo) GetServices(ctx context.Context, employeeID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, employeeID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// nextBusinessSlot returns a start time that passes booking validation:
// tomorrow at 10:00 local time, aligned to the 30-minute slot grid.
func nextBusinessSlot() time.Time {
	tomorrow := time.Now().Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.Local)
}

func TestBookingCreate_Success(t *testing.T) {
	bRepo := new(mockBookingRepo)
	sRepo := new(mockServiceRepo)
	eRepo := new(mockEmployeeRepo)
	svc := service.NewBookingService(bRepo, eRepo, sRepo)

	svcID := uuid.New()
	empID := uuid.New()
	userID := uuid.New()
	start := nextBusinessSlot()

	sRepo.On("GetByID", mock.Anything, svcID).Return(&domain.Service{
		ID: svcID, Name: "Haircut", Duration: 30, Price: 50,
	}, nil)

	eRepo.On("GetByID", mock.Anything, empID).Return(&domain.Employee{
		ID: empID, UserID: uuid.New(), Bio: "Stylist",
	}, nil)

	bRepo.On("CreateAtomic", mock.Anything,
		mock.MatchedBy(func(b *domain.Booking) bool {
			return b.UserID == userID && b.EmployeeID == empID && b.ServiceID == svcID
		}),
		empID, mock.Anything, mock.Anything,
	).Return(nil)

	booking, err := svc.Create(context.Background(), userID, empID, svcID, start)
	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, domain.StatusConfirmed, booking.Status)
	assert.Equal(t, 50.0, booking.TotalPrice)
	bRepo.AssertExpectations(t)
	sRepo.AssertExpectations(t)
	eRepo.AssertExpectations(t)
}

func TestBookingCreate_Conflict(t *testing.T) {
	bRepo := new(mockBookingRepo)
	sRepo := new(mockServiceRepo)
	eRepo := new(mockEmployeeRepo)
	svc := service.NewBookingService(bRepo, eRepo, sRepo)

	svcID := uuid.New()
	empID := uuid.New()
	userID := uuid.New()
	start := nextBusinessSlot()

	sRepo.On("GetByID", mock.Anything, svcID).Return(&domain.Service{
		ID: svcID, Duration: 30,
	}, nil)

	eRepo.On("GetByID", mock.Anything, empID).Return(&domain.Employee{
		ID: empID, UserID: uuid.New(),
	}, nil)

	bRepo.On("CreateAtomic", mock.Anything, mock.Anything,
		empID, mock.Anything, mock.Anything,
	).Return(repository.ErrConflict)

	booking, err := svc.Create(context.Background(), userID, empID, svcID, start)
	assert.Error(t, err)
	assert.Equal(t, service.ErrConflict, err)
	assert.Nil(t, booking)
	bRepo.AssertExpectations(t)
	sRepo.AssertExpectations(t)
	eRepo.AssertExpectations(t)
}

func TestBookingCreate_StartTimeValidation(t *testing.T) {
	cases := []struct {
		name    string
		start   time.Time
		wantErr error
	}{
		{"past", nextBusinessSlot().Add(-48 * time.Hour), service.ErrPastStartTime},
		{"unaligned", nextBusinessSlot().Add(7 * time.Minute), service.ErrUnalignedStartTime},
		{"before opening", nextBusinessSlot().Add(-7 * time.Hour), service.ErrOutsideHours},                  // 03:00
		{"runs past closing", nextBusinessSlot().Add(7*time.Hour + 30*time.Minute), service.ErrOutsideHours}, // 17:30 + 60min
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bRepo := new(mockBookingRepo)
			sRepo := new(mockServiceRepo)
			eRepo := new(mockEmployeeRepo)
			svc := service.NewBookingService(bRepo, eRepo, sRepo)

			svcID := uuid.New()
			empID := uuid.New()

			sRepo.On("GetByID", mock.Anything, svcID).Return(&domain.Service{
				ID: svcID, Duration: 60,
			}, nil)
			eRepo.On("GetByID", mock.Anything, empID).Return(&domain.Employee{
				ID: empID,
			}, nil)

			booking, err := svc.Create(context.Background(), uuid.New(), empID, svcID, tc.start)
			assert.Equal(t, tc.wantErr, err)
			assert.Nil(t, booking)
			bRepo.AssertNotCalled(t, "CreateAtomic")
		})
	}
}

func TestBookingGetByID_OwnershipEnforced(t *testing.T) {
	bRepo := new(mockBookingRepo)
	sRepo := new(mockServiceRepo)
	eRepo := new(mockEmployeeRepo)
	svc := service.NewBookingService(bRepo, eRepo, sRepo)

	ownerID := uuid.New()
	otherID := uuid.New()
	bookingID := uuid.New()

	bRepo.On("GetByID", mock.Anything, bookingID).Return(&domain.Booking{
		ID: bookingID, UserID: ownerID,
	}, nil)

	// Owner can read.
	b, err := svc.GetByID(context.Background(), bookingID, ownerID, "user")
	assert.NoError(t, err)
	assert.NotNil(t, b)

	// A different non-admin user cannot.
	b, err = svc.GetByID(context.Background(), bookingID, otherID, "user")
	assert.Equal(t, service.ErrNotAuthorized, err)
	assert.Nil(t, b)

	// An admin can.
	b, err = svc.GetByID(context.Background(), bookingID, otherID, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
