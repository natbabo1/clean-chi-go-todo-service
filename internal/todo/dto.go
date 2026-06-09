package todo

import (
	"time"

	"github.com/google/uuid"
)

type CreateInput struct {
	Title       string     `json:"title"       validate:"required"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Completed   *bool      `json:"completed"`
	DueDate     *time.Time `json:"due_date"`
}

type ListParams struct {
	UserID    uuid.UUID
	Completed *bool
	Page      int
	Limit     int
}

// Response is the public representation of a todo.
type Response struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Completed   bool       `json:"completed"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToResponse(t *Todo) Response {
	return Response{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
