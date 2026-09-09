package auth_test

import (
	"testing"

	"github.com/nixwai/go-game-server/app/model"
	authdto "github.com/nixwai/go-game-server/app/module/auth/dto"
)

func TestNewUserResponseDoesNotExposePasswordHash(t *testing.T) {
	user := model.User{ID: 1, Username: "alice", PasswordHash: "secret-hash", Role: model.RoleUser, Status: model.StatusActive}
	got := authdto.NewUserResponse(user)
	if got.ID != user.ID || got.Username != user.Username || got.Role != user.Role || got.Status != user.Status {
		t.Fatalf("unexpected user response: %+v", got)
	}
}

func TestNewLoginResponseMapsTokenAndUser(t *testing.T) {
	user := model.User{ID: 1, Username: "alice", PasswordHash: "secret-hash", Role: model.RoleUser, Status: model.StatusActive}
	got := authdto.NewLoginResponse("token", user)
	if got.Token != "token" || got.User.Username != user.Username || got.User.Role != user.Role {
		t.Fatalf("unexpected login response: %+v", got)
	}
}
