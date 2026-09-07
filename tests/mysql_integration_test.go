//go:build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/nixwai/go-game-server/app/database"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
)

func TestMySQLUserRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}
	db, err := database.OpenMySQL(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewGormUserRepository(db)
	user := model.User{Username: "integration_user", PasswordHash: "argon2id-test", Role: model.RoleUser, Status: model.StatusActive}
	if err := repo.Create(context.Background(), &user); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByUsername(context.Background(), user.Username)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 || got.Username != user.Username {
		t.Fatalf("unexpected user: %+v", got)
	}
	db.Exec("DELETE FROM users WHERE username = ?", user.Username)
}
