package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/security"
	"github.com/nixwai/go-game-server/app/service"
)

type memoryUsers struct {
	byName    map[string]model.User
	next      uint64
	err       error
	createErr error
}

func newMemoryUsers() *memoryUsers { return &memoryUsers{byName: map[string]model.User{}, next: 1} }
func (m *memoryUsers) FindByUsername(_ context.Context, username string) (model.User, error) {
	if m.err != nil {
		return model.User{}, m.err
	}
	user, ok := m.byName[username]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return user, nil
}
func (m *memoryUsers) FindByID(_ context.Context, id uint64) (model.User, error) {
	for _, user := range m.byName {
		if user.ID == id {
			return user, nil
		}
	}
	return model.User{}, repository.ErrNotFound
}
func (m *memoryUsers) Create(_ context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, ok := m.byName[user.Username]; ok {
		return repository.ErrDuplicate
	}
	user.ID = m.next
	m.next++
	m.byName[user.Username] = *user
	return nil
}
func newTestService(users *memoryUsers) *service.AuthService {
	hasher := security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
	tokens := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	return service.NewAuthService(users, hasher, tokens)
}
func TestRegisterAndLogin(t *testing.T) {
	users := newMemoryUsers()
	service := newTestService(users)
	user, err := service.Register(context.Background(), "alice", "SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != model.RoleUser || user.Username != "alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
	result, err := service.Login(context.Background(), "alice", "SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	if result.Token == "" {
		t.Fatal("token should be returned")
	}
}
func TestRegisterRejectsInvalidInputAndDuplicate(t *testing.T) {
	service := newTestService(newMemoryUsers())
	if _, err := service.Register(context.Background(), "ab", "SecurePass123"); err == nil {
		t.Fatal("invalid username should fail")
	}
	if _, err := service.Register(context.Background(), "alice", "weakpass"); err == nil {
		t.Fatal("invalid password should fail")
	}
	if _, err := service.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(context.Background(), "alice", "SecurePass123"); err == nil {
		t.Fatal("duplicate should fail")
	}
}
func TestLoginFailureForMissingAndWrongPassword(t *testing.T) {
	service := newTestService(newMemoryUsers())
	if _, err := service.Login(context.Background(), "missing", "SecurePass123"); err == nil {
		t.Fatal("missing user login should fail")
	}
	if _, err := service.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(context.Background(), "alice", "WrongPass123"); err == nil {
		t.Fatal("wrong password should fail")
	}
}
