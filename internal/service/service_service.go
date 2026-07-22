package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) Create(ctx context.Context, name, description string, duration int, price float64) (*domain.Service, error) {
	now := time.Now()
	svc := &domain.Service{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Duration:    duration,
		Price:       price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) List(ctx context.Context) ([]domain.Service, error) {
	return s.repo.List(ctx)
}

func (s *ServiceService) Update(ctx context.Context, id uuid.UUID, name, description string, duration int, price float64) (*domain.Service, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	svc.Name = name
	svc.Description = description
	svc.Duration = duration
	svc.Price = price
	svc.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrServiceNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}
