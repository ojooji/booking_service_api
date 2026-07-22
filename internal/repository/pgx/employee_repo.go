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

type employeeRepo struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepository(pool *pgxpool.Pool) *employeeRepo {
	return &employeeRepo{pool: pool}
}

func (r *employeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	query := `INSERT INTO employees (id, user_id, bio, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, emp.ID, emp.UserID, emp.Bio, emp.CreatedAt, emp.UpdatedAt)
	return err
}

func (r *employeeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	query := `SELECT id, user_id, bio, created_at, updated_at FROM employees WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanEmployee(row)
}

func (r *employeeRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error) {
	query := `SELECT id, user_id, bio, created_at, updated_at FROM employees WHERE user_id = $1`
	row := r.pool.QueryRow(ctx, query, userID)
	return scanEmployee(row)
}

func (r *employeeRepo) List(ctx context.Context) ([]domain.Employee, error) {
	query := `SELECT id, user_id, bio, created_at, updated_at FROM employees ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployees(rows)
}

func (r *employeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	query := `UPDATE employees SET bio=$1, updated_at=$2 WHERE id=$3`
	_, err := r.pool.Exec(ctx, query, emp.Bio, emp.UpdatedAt, emp.ID)
	return err
}

func (r *employeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op if already committed

	if _, err := tx.Exec(ctx, `DELETE FROM employee_services WHERE employee_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM employees WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *employeeRepo) SetServices(ctx context.Context, employeeID uuid.UUID, serviceIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op if already committed

	if _, err := tx.Exec(ctx, `DELETE FROM employee_services WHERE employee_id = $1`, employeeID); err != nil {
		return err
	}
	for _, sid := range serviceIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO employee_services (employee_id, service_id) VALUES ($1, $2)`, employeeID, sid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *employeeRepo) GetServices(ctx context.Context, employeeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT service_id FROM employee_services WHERE employee_id = $1`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func scanEmployee(row pgx.Row) (*domain.Employee, error) {
	e := &domain.Employee{}
	err := row.Scan(&e.ID, &e.UserID, &e.Bio, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return e, err
}

func scanEmployees(rows pgx.Rows) ([]domain.Employee, error) {
	var emps []domain.Employee
	for rows.Next() {
		var e domain.Employee
		if err := rows.Scan(&e.ID, &e.UserID, &e.Bio, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		emps = append(emps, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return emps, nil
}
