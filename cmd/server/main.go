// Command server 启动 Go Game Server HTTP 服务。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/database"
	"github.com/nixwai/go-game-server/app/handler"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/router"
	"github.com/nixwai/go-game-server/app/security"
	"github.com/nixwai/go-game-server/app/service"
)

// main 加载配置、组装依赖并以优雅停机方式运行 HTTP 服务。
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := database.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("get database", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	// 在启动入口集中完成依赖组装，保持各层通过接口协作。
	users := repository.NewGormUserRepository(db)
	hasher := security.PasswordHasher{Time: cfg.Argon2Time, Memory: cfg.Argon2Memory, Threads: cfg.Argon2Threads, KeyLen: cfg.Argon2KeyLen, SaltLen: cfg.Argon2SaltLen}
	tokens := security.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiresIn)
	authService := service.NewAuthService(users, hasher, tokens)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router.New(handler.NewAuthHandler(authService), tokens),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// 阻塞等待操作系统终止信号，避免服务被直接杀死而丢失未完成请求。
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
}
