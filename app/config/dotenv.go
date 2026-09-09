package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotEnv 从当前工作目录加载 .env，同时保留已注入的进程环境变量。
func LoadDotEnv() error {
	err := godotenv.Load()
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("load .env: %w", err)
}
