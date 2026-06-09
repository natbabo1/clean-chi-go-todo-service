package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-list/internal/auth"
	"github.com/user/todo-list/internal/platform/jwt"
	"github.com/user/todo-list/internal/user"
)

// fakeUserRepo implements user.Repository in memory.
type fakeUserRepo struct {
	users map[string]*user.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*user.User)}
}

func (r *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	r.users[u.Email] = u
	return nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	u, ok := r.users[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}

// fakeHasher stores plaintext for simplicity in tests.
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error)        { return "hashed:" + plain, nil }
func (fakeHasher) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return errors.New("wrong password")
	}
	return nil
}

func newTestService(repo user.Repository) auth.Service {
	mgr := jwt.NewManager("test-secret-key-that-is-long-enough", 24*time.Hour)
	return auth.NewService(repo, fakeHasher{}, mgr)
}

func TestRegister_Success(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	res, err := svc.Register(context.Background(), auth.RegisterInput{
		Email:    "alice@example.com",
		Password: "password123",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if res.User.Email != "alice@example.com" {
		t.Fatalf("wrong email: %s", res.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	in := auth.RegisterInput{Email: "dup@example.com", Password: "password123", Name: "Dup"}
	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Register(context.Background(), in)
	if !errors.Is(err, auth.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Email: "bob@example.com", Password: "secret", Name: "Bob",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "bob@example.com", Password: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.AccessToken == "" {
		t.Fatal("expected token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	_, _ = svc.Register(context.Background(), auth.RegisterInput{
		Email: "carol@example.com", Password: "correct", Name: "Carol",
	})

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "carol@example.com", Password: "wrong",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
