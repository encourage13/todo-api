package model

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Completed   bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
