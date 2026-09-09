// Package bootstrap 负责构建并运行 HTTP 服务应用。
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/handler"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/router"
	"github.com/nixwai/go-game-server/app/security"
	"github.com/nixwai/go-game-server/app/service"
)

const shutdownTimeout = 10 * time.Second

// HTTPServer 定义 HTTP 服务生命周期所需的最小接口，便于隔离启动与关闭逻辑。
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// Application 持有 HTTP 服务并负责管理其生命周期。
type Application struct {
	httpServer *http.Server
}

// New 根据 DB 和配置创建 HTTP 服务应用并组装全部运行时依赖。
func New(cfg config.Config, db *DB) (*Application, error) {
	authHandler, aiHandler, tokens, err := buildDependencies(cfg, db)
	if err != nil {
		return nil, err
	}
	return &Application{
		httpServer: buildHTTPServer(cfg, authHandler, aiHandler, tokens),
	}, nil
}

// buildDependencies 构建认证和 AI 管理相关的 Repository、Service、Handler 和 Token 管理器。
func buildDependencies(cfg config.Config, db *DB) (*handler.AuthHandler, *handler.AIHandler, *security.TokenManager, error) {
	users := repository.NewGormUserRepository(db.GORM)
	hasher := security.PasswordHasher{
		Time:    cfg.Argon2Time,
		Memory:  cfg.Argon2Memory,
		Threads: cfg.Argon2Threads,
		KeyLen:  cfg.Argon2KeyLen,
		SaltLen: cfg.Argon2SaltLen,
	}
	tokens := security.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiresIn)
	authService := service.NewAuthService(users, hasher, tokens)
	authHandler := handler.NewAuthHandler(authService)

	// 构建加密管理器和 AI 管理服务。
	crypto, err := security.NewCryptoManager(cfg.MasterKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create crypto manager: %w", err)
	}
	aiProviders := repository.NewGormAIProviderRepository(db.GORM)
	aiModels := repository.NewGormAIModelRepository(db.GORM)
	aiService := service.NewAIService(aiProviders, aiModels, crypto, cfg.DefaultAI)
	aiHandler := handler.NewAIHandler(aiService, crypto)

	return authHandler, aiHandler, tokens, nil
}

// buildHTTPServer 根据配置创建带超时控制的 HTTP 服务器。
func buildHTTPServer(cfg config.Config, authHandler *handler.AuthHandler, aiHandler *handler.AIHandler, tokens *security.TokenManager) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router.New(authHandler, aiHandler, tokens),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Run 启动 HTTP 服务，在上下文结束时优雅停机。
func (a *Application) Run(ctx context.Context) error {
	if a == nil || a.httpServer == nil {
		return errors.New("HTTP server is not initialized")
	}
	return RunHTTPServer(ctx, a.httpServer, shutdownTimeout)
}

// RunHTTPServer 启动 HTTP 服务；收到上下文结束信号后执行优雅停机。
func RunHTTPServer(ctx context.Context, server HTTPServer, shutdownTimeout time.Duration) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if server == nil {
		return errors.New("HTTP server is nil")
	}
	if shutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("HTTP server stopped unexpectedly: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := server.Shutdown(shutdownCtx)
		serveErr := <-serveErr
		if shutdownErr != nil {
			return fmt.Errorf("shutdown HTTP server: %w", shutdownErr)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server stopped during shutdown: %w", serveErr)
		}
		return nil
	}
}
