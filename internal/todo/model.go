package todo

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description *string
	Completed   bool
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
