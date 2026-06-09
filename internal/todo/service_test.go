package todo_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/user/todo-list/internal/todo"
)

type fakeTodoRepo struct {
	mu    sync.Mutex
	todos map[uuid.UUID]*todo.Todo
}

func newFakeTodoRepo() *fakeTodoRepo {
	return &fakeTodoRepo{todos: make(map[uuid.UUID]*todo.Todo)}
}

func (r *fakeTodoRepo) Create(_ context.Context, t *todo.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.todos[t.ID] = t
	return nil
}

func (r *fakeTodoRepo) FindByID(_ context.Context, userID, todoID uuid.UUID) (*todo.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.todos[todoID]
	if !ok || t.UserID != userID {
		return nil, todo.ErrNotFound
	}
	return t, nil
}

func (r *fakeTodoRepo) List(_ context.Context, p todo.ListParams) ([]todo.Todo, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []todo.Todo
	for _, t := range r.todos {
		if t.UserID != p.UserID {
			continue
		}
		if p.Completed != nil && t.Completed != *p.Completed {
			continue
		}
		out = append(out, *t)
	}
	return out, len(out), nil
}

func (r *fakeTodoRepo) Update(_ context.Context, t *todo.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.todos[t.ID]; !ok {
		return todo.ErrNotFound
	}
	r.todos[t.ID] = t
	return nil
}

func (r *fakeTodoRepo) Delete(_ context.Context, userID, todoID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.todos[todoID]
	if !ok || t.UserID != userID {
		return todo.ErrNotFound
	}
	delete(r.todos, todoID)
	return nil
}

func TestCreate_Success(t *testing.T) {
	svc := todo.NewService(newFakeTodoRepo())
	userID := uuid.New()

	got, err := svc.Create(context.Background(), userID, todo.CreateInput{Title: "Buy milk"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Buy milk" {
		t.Fatalf("wrong title: %s", got.Title)
	}
	if got.UserID != userID {
		t.Fatal("wrong user_id on created todo")
	}
}

func TestList_ScopedToUser(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := todo.NewService(repo)

	alice := uuid.New()
	bob := uuid.New()

	_, _ = svc.Create(context.Background(), alice, todo.CreateInput{Title: "Alice todo"})
	_, _ = svc.Create(context.Background(), bob, todo.CreateInput{Title: "Bob todo"})

	aliceTodos, total, err := svc.List(context.Background(), todo.ListParams{UserID: alice, Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(aliceTodos) != 1 {
		t.Fatalf("expected 1 todo for alice, got %d", total)
	}
	if aliceTodos[0].UserID != alice {
		t.Fatal("todo belongs to wrong user")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := todo.NewService(newFakeTodoRepo())

	_, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetByID_OtherUserCannotAccess(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := todo.NewService(repo)

	owner := uuid.New()
	attacker := uuid.New()

	created, _ := svc.Create(context.Background(), owner, todo.CreateInput{Title: "Private"})

	_, err := svc.GetByID(context.Background(), attacker, created.ID)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when accessing other user's todo, got %v", err)
	}
}
