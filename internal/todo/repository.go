package todo

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, t *Todo) error
	FindByID(ctx context.Context, userID, todoID uuid.UUID) (*Todo, error)
	List(ctx context.Context, p ListParams) ([]Todo, int, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, userID, todoID uuid.UUID) error
}
