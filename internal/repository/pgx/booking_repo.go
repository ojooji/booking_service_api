package pgx

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ojooji/booking-service-api/internal/domain"
	"github.com/ojooji/booking-service-api/internal/repository"
)

type bookingRepo struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *bookingRepo {
	return &bookingRepo{pool: pool}
}

func (r *bookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, user_id, employee_id, service_id, start_time, end_time, status, total_price, created_at, updated_at FROM bookings WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanBooking(row)
}

func (r *bookingRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	query := `SELECT id, user_id, employee_id, service_id, start_time, end_time, status, total_price, created_at, updated_at FROM bookings WHERE user_id = $1 ORDER BY start_time DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBookings(rows)
}

func (r *bookingRepo) ListByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]domain.Booking, error) {
	query := `SELECT id, user_id, employee_id, service_id, start_time, end_time, status, total_price, created_at, updated_at
		FROM bookings
		WHERE employee_id = $1
		AND status != 'cancelled'
		AND start_time < $3
		AND end_time > $2
		ORDER BY start_time`
	rows, err := r.pool.Query(ctx, query, employeeID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBookings(rows)
}

func (r *bookingRepo) ListAll(ctx context.Context) ([]domain.Booking, error) {
	query := `SELECT id, user_id, employee_id, service_id, start_time, end_time, status, total_price, created_at, updated_at FROM bookings ORDER BY start_time DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBookings(rows)
}

func (r *bookingRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE bookings SET status=$1, updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	return err
}

func (r *bookingRepo) CreateAtomic(ctx context.Context, booking *domain.Booking, employeeID uuid.UUID, start, end time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op if already committed

	// Serialize concurrent booking attempts for the same employee so the
	// conflict check below cannot race under READ COMMITTED. The lock is
	// released automatically at commit/rollback.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1::text, 0))`, employeeID); err != nil {
		return err
	}

	var count int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM bookings
		WHERE employee_id = $1
		AND status != 'cancelled'
		AND start_time < $3
		AND end_time > $2`,
		employeeID, start, end).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return repository.ErrConflict
	}

	query := `INSERT INTO bookings (id, user_id, employee_id, service_id, start_time, end_time, status, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(ctx, query, booking.ID, booking.UserID, booking.EmployeeID, booking.ServiceID,
		booking.StartTime, booking.EndTime, booking.Status, booking.TotalPrice, booking.CreatedAt, booking.UpdatedAt)
	if err != nil {
		// The bookings_no_overlap exclusion constraint is the last line of
		// defense; surface its violation as a conflict, not a 500.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return repository.ErrConflict
		}
		return err
	}

	return tx.Commit(ctx)
}

func scanBooking(row pgx.Row) (*domain.Booking, error) {
	b := &domain.Booking{}
	err := row.Scan(&b.ID, &b.UserID, &b.EmployeeID, &b.ServiceID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return b, err
}

func scanBookings(rows pgx.Rows) ([]domain.Booking, error) {
	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.EmployeeID, &b.ServiceID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}
