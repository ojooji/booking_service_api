package dto

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyEmail      = errors.New("email is required")
	ErrEmptyPassword   = errors.New("password is required")
	ErrShortPassword   = errors.New("password must be at least 8 characters")
	ErrEmptyName       = errors.New("first and last name are required")
	ErrInvalidPrice    = errors.New("price must be non-negative")
	ErrInvalidDuration = errors.New("duration must be positive")
)

func ValidateRegister(r RegisterRequest) error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrEmptyEmail
	}
	if r.Password == "" {
		return ErrEmptyPassword
	}
	if len(r.Password) < 8 {
		return ErrShortPassword
	}
	if strings.TrimSpace(r.FirstName) == "" || strings.TrimSpace(r.LastName) == "" {
		return ErrEmptyName
	}
	return nil
}

func ValidateCreateService(r CreateServiceRequest) error {
	if r.Duration <= 0 {
		return ErrInvalidDuration
	}
	if r.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

func ValidateUpdateService(r UpdateServiceRequest) error {
	if r.Duration <= 0 {
		return ErrInvalidDuration
	}
	if r.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
}

type CreateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	Price       float64 `json:"price"`
}

type UpdateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	Price       float64 `json:"price"`
}

type CreateEmployeeRequest struct {
	UserID     string      `json:"user_id"`
	Bio        string      `json:"bio"`
	ServiceIDs []uuid.UUID `json:"service_ids"`
}

type UpdateEmployeeRequest struct {
	Bio        string      `json:"bio"`
	ServiceIDs []uuid.UUID `json:"service_ids"`
}

type CreateBookingRequest struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	StartTime  time.Time `json:"start_time"`
}

type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
