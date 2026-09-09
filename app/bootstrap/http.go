// Package bootstrap 负责构建并运行 HTTP 服务应用。
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nixwai/go-game-server/app/config"
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

// New 根据 HTTP 监听配置和路由处理器创建 HTTP 服务应用。
func New(cfg config.Config, handler http.Handler) *Application {
	return &Application{
		httpServer: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
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
