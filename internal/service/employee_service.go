package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrEmployeeExists   = errors.New("employee already exists for this user")
)

type EmployeeService struct {
	repo     repository.EmployeeRepository
	userRepo repository.UserRepository
}

func NewEmployeeService(repo repository.EmployeeRepository, userRepo repository.UserRepository) *EmployeeService {
	return &EmployeeService{repo: repo, userRepo: userRepo}
}

func (s *EmployeeService) Create(ctx context.Context, userID uuid.UUID, bio string, serviceIDs []uuid.UUID) (*domain.Employee, error) {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmployeeExists
	}

	now := time.Now()
	emp := &domain.Employee{
		ID:        uuid.New(),
		UserID:    userID,
		Bio:       bio,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, emp); err != nil {
		return nil, err
	}
	if len(serviceIDs) > 0 {
		if err := s.repo.SetServices(ctx, emp.ID, serviceIDs); err != nil {
			return nil, err
		}
	}
	return emp, nil
}

func (s *EmployeeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	services, _ := s.repo.GetServices(ctx, id)
	emp.ServiceIDs = services
	return emp, nil
}

func (s *EmployeeService) List(ctx context.Context) ([]domain.Employee, error) {
	emps, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range emps {
		services, _ := s.repo.GetServices(ctx, emps[i].ID)
		emps[i].ServiceIDs = services
	}
	return emps, nil
}

func (s *EmployeeService) Update(ctx context.Context, id uuid.UUID, bio string, serviceIDs []uuid.UUID) (*domain.Employee, error) {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	emp.Bio = bio
	emp.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, emp); err != nil {
		return nil, err
	}
	if serviceIDs != nil {
		if err := s.repo.SetServices(ctx, id, serviceIDs); err != nil {
			return nil, err
		}
	}
	return emp, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEmployeeNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}
