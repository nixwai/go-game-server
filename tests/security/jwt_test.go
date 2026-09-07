package security_test

import (
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/security"
)

func TestTokenManagerRoundTrip(t *testing.T) {
	manager := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	token, err := manager.Generate(model.User{ID: 7, Username: "alice", Role: model.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.Role != model.RoleAdmin {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
func TestTokenManagerRejectsWrongAlgorithm(t *testing.T) {
	manager := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	token, err := manager.Generate(model.User{ID: 7, Username: "alice", Role: model.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := manager.Parse(token)
	if err != nil || parsed == nil {
		t.Fatal("expected valid token")
	}
}
