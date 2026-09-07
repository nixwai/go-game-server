package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
)

func TestCurrentUserAndRepositoryErrors(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	if _, err := svc.CurrentUser(context.Background(), 1); err == nil {
		t.Fatal("missing user should fail")
	}
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	user, err := svc.CurrentUser(context.Background(), 1)
	if err != nil || user.Username != "alice" {
		t.Fatalf("unexpected current user: %+v, %v", user, err)
	}
	users.createErr = errors.New("database unavailable")
	if _, err := svc.Register(context.Background(), "bob", "SecurePass123"); err == nil {
		t.Fatal("repository error should fail")
	}
}
func TestDisabledUserCannotLogin(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	_, err := svc.Register(context.Background(), "alice", "SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	user := users.byName["alice"]
	user.Status = model.StatusDisabled
	users.byName["alice"] = user
	if _, err := svc.Login(context.Background(), "alice", "SecurePass123"); err == nil {
		t.Fatal("disabled user should fail")
	}
	_ = repository.ErrNotFound
}
