package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nixwai/go-game-server/app/config"
)

func TestLoadDotEnvLoadsFileWithoutOverwritingProcessEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	dotEnvPath := filepath.Join(tempDir, ".env")
	content := "DOTENV_FILE_VALUE=from-file\nDOTENV_PROCESS_VALUE=from-file\n"
	if err := os.WriteFile(dotEnvPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	unsetEnvForTest(t, "DOTENV_FILE_VALUE")
	t.Setenv("DOTENV_PROCESS_VALUE", "from-process")
	t.Chdir(tempDir)

	if err := config.LoadDotEnv(); err != nil {
		t.Fatal(err)
	}
	if value := os.Getenv("DOTENV_FILE_VALUE"); value != "from-file" {
		t.Fatalf("expected .env value, got %q", value)
	}
	if value := os.Getenv("DOTENV_PROCESS_VALUE"); value != "from-process" {
		t.Fatalf("expected process environment to win, got %q", value)
	}
}

func TestLoadDotEnvAllowsMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := config.LoadDotEnv(); err != nil {
		t.Fatalf("missing .env should be allowed: %v", err)
	}
}

func TestLoadDotEnvRejectsInvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte("INVALID LINE\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tempDir)

	if err := config.LoadDotEnv(); err == nil {
		t.Fatal("expected invalid .env error")
	}
}

func TestLoadAutomaticallyReadsDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	dotEnvPath := filepath.Join(tempDir, ".env")
	content := "MYSQL_DSN=user:pass@tcp(localhost:3306)/db\n" +
		"JWT_SECRET=01234567890123456789012345678901\n"
	if err := os.WriteFile(dotEnvPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	unsetEnvForTest(t, "MYSQL_DSN")
	unsetEnvForTest(t, "JWT_SECRET")
	t.Chdir(tempDir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MySQLDSN != "user:pass@tcp(localhost:3306)/db" {
		t.Fatalf("unexpected MySQL DSN: %q", cfg.MySQLDSN)
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	value, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}
