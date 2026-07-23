package domain

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	Bio        string      `json:"bio"`
	ServiceIDs []uuid.UUID `json:"service_ids"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
