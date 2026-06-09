package todo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, t *Todo) error {
	const q = `
		INSERT INTO todos (id, user_id, title, description, completed, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q, t.ID, t.UserID, t.Title, t.Description, t.Completed, t.DueDate, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert todo: %w", err)
	}
	return nil
}

func (r *postgresRepository) FindByID(ctx context.Context, userID, todoID uuid.UUID) (*Todo, error) {
	const q = `
		SELECT id, user_id, title, description, completed, due_date, created_at, updated_at
		FROM todos WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRow(ctx, q, todoID, userID)
	t, err := scanTodo(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *postgresRepository) List(ctx context.Context, p ListParams) ([]Todo, int, error) {
	offset := (p.Page - 1) * p.Limit

	// Build query dynamically based on optional completed filter.
	baseWhere := `WHERE user_id = $1`
	args := []any{p.UserID}
	if p.Completed != nil {
		baseWhere += fmt.Sprintf(` AND completed = $%d`, len(args)+1)
		args = append(args, *p.Completed)
	}

	countQ := `SELECT COUNT(*) FROM todos ` + baseWhere
	var total int
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count todos: %w", err)
	}

	listQ := `SELECT id, user_id, title, description, completed, due_date, created_at, updated_at
		FROM todos ` + baseWhere +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, p.Limit, offset)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, 0, err
		}
		todos = append(todos, *t)
	}
	return todos, total, rows.Err()
}

func (r *postgresRepository) Update(ctx context.Context, t *Todo) error {
	const q = `
		UPDATE todos SET title=$1, description=$2, completed=$3, due_date=$4, updated_at=$5
		WHERE id=$6 AND user_id=$7`
	tag, err := r.db.Exec(ctx, q, t.Title, t.Description, t.Completed, t.DueDate, t.UpdatedAt, t.ID, t.UserID)
	if err != nil {
		return fmt.Errorf("update todo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, userID, todoID uuid.UUID) error {
	const q = `DELETE FROM todos WHERE id = $1 AND user_id = $2`
	tag, err := r.db.Exec(ctx, q, todoID, userID)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanTodo(row scannable) (*Todo, error) {
	var t Todo
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan todo: %w", err)
	}
	return &t, nil
}
