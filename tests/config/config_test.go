package config_test

import (
	"testing"

	"github.com/nixwai/go-game-server/app/config"
)

func TestLoadRejectsMissingRequiredValues(t *testing.T) {
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("JWT_SECRET", "short")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected missing MySQL error")
	}
}

func TestLoadAndAdminLoad(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/db")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.JWTExpiresIn.String() != "2h0m0s" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	t.Setenv("JWT_SECRET", "")
	if _, err := config.LoadForAdmin(); err != nil {
		t.Fatal(err)
	}
}
