package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/model"
	auth "github.com/nixwai/go-game-server/app/module/auth"

	"github.com/nixwai/go-game-server/app/security"
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
		return model.User{}, model.ErrNotFound
	}
	return user, nil
}
func (m *memoryUsers) FindByID(_ context.Context, id uint64) (model.User, error) {
	for _, user := range m.byName {
		if user.ID == id {
			return user, nil
		}
	}
	return model.User{}, model.ErrNotFound
}
func (m *memoryUsers) Create(_ context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, ok := m.byName[user.Username]; ok {
		return model.ErrDuplicate
	}
	user.ID = m.next
	m.next++
	m.byName[user.Username] = *user
	return nil
}
func (m *memoryUsers) UpdatePassword(_ context.Context, id uint64, passwordHash string) error {
	for name, user := range m.byName {
		if user.ID == id {
			user.PasswordHash = passwordHash
			m.byName[name] = user
			return nil
		}
	}
	return model.ErrNotFound
}

func newTestService(users *memoryUsers) *auth.Service {
	hasher := security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
	tokens := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	return auth.NewService(users, hasher, tokens)
}

func TestRegisterAndLogin(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	user, err := svc.Register(context.Background(), "alice", "SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != model.RoleUser || user.Username != "alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
	result, err := svc.Login(context.Background(), "alice", "SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	if result.Token == "" {
		t.Fatal("token should be returned")
	}
}

func TestRegisterRejectsInvalidInputAndDuplicate(t *testing.T) {
	svc := newTestService(newMemoryUsers())
	if _, err := svc.Register(context.Background(), "ab", "SecurePass123"); err == nil {
		t.Fatal("invalid username should fail")
	}
	if _, err := svc.Register(context.Background(), "alice", "weakpass"); err == nil {
		t.Fatal("invalid password should fail")
	}
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err == nil {
		t.Fatal("duplicate should fail")
	}
}

func TestLoginFailureForMissingAndWrongPassword(t *testing.T) {
	svc := newTestService(newMemoryUsers())
	if _, err := svc.Login(context.Background(), "missing", "SecurePass123"); err == nil {
		t.Fatal("missing user login should fail")
	}
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login(context.Background(), "alice", "WrongPass123"); err == nil {
		t.Fatal("wrong password should fail")
	}
}

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
}

func TestChangePasswordSuccess(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ChangePassword(context.Background(), 1, "SecurePass123", "NewPass456"); err != nil {
		t.Fatalf("change password failed: %v", err)
	}
	if _, err := svc.Login(context.Background(), "alice", "SecurePass123"); err == nil {
		t.Fatal("old password should fail after change")
	}
	result, err := svc.Login(context.Background(), "alice", "NewPass456")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	if result.Token == "" {
		t.Fatal("token should be returned")
	}
}

func TestChangePasswordWrongOldPassword(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ChangePassword(context.Background(), 1, "WrongPass123", "NewPass456"); err == nil {
		t.Fatal("wrong old password should fail")
	}
}

func TestChangePasswordInvalidNewPassword(t *testing.T) {
	users := newMemoryUsers()
	svc := newTestService(users)
	if _, err := svc.Register(context.Background(), "alice", "SecurePass123"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ChangePassword(context.Background(), 1, "SecurePass123", "weak"); err == nil {
		t.Fatal("invalid new password should fail")
	}
}

func TestChangePasswordUserNotFound(t *testing.T) {
	svc := newTestService(newMemoryUsers())
	if err := svc.ChangePassword(context.Background(), 999, "SecurePass123", "NewPass456"); err == nil {
		t.Fatal("nonexistent user should fail")
	}
}
