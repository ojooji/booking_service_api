package pgx

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
)

type serviceRepo struct {
	pool *pgxpool.Pool
}

func NewServiceRepository(pool *pgxpool.Pool) *serviceRepo {
	return &serviceRepo{pool: pool}
}

func (r *serviceRepo) Create(ctx context.Context, svc *domain.Service) error {
	query := `INSERT INTO services (id, name, description, duration, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query, svc.ID, svc.Name, svc.Description, svc.Duration, svc.Price, svc.CreatedAt, svc.UpdatedAt)
	return err
}

func (r *serviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	query := `SELECT id, name, description, duration, price, created_at, updated_at FROM services WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanService(row)
}

func (r *serviceRepo) List(ctx context.Context) ([]domain.Service, error) {
	query := `SELECT id, name, description, duration, price, created_at, updated_at FROM services ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServices(rows)
}

func (r *serviceRepo) Update(ctx context.Context, svc *domain.Service) error {
	query := `UPDATE services SET name=$1, description=$2, duration=$3, price=$4, updated_at=$5 WHERE id=$6`
	_, err := r.pool.Exec(ctx, query, svc.Name, svc.Description, svc.Duration, svc.Price, svc.UpdatedAt, svc.ID)
	return err
}

func (r *serviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	return err
}

func scanService(row pgx.Row) (*domain.Service, error) {
	s := &domain.Service{}
	err := row.Scan(&s.ID, &s.Name, &s.Description, &s.Duration, &s.Price, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return s, err
}

func scanServices(rows pgx.Rows) ([]domain.Service, error) {
	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Duration, &s.Price, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return services, nil
}
