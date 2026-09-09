// Command server 启动 Go Game Server HTTP 服务。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nixwai/go-game-server/app/bootstrap"
	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/router"
)

// main 加载配置、连接数据库、组装应用并运行 HTTP 服务，统一处理启动、运行和关闭错误。
func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(fmt.Errorf("load config: %w", err))
	}
	db, err := bootstrap.OpenDB(cfg.MySQLDSN)
	if err != nil {
		fatal(err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			fatal(closeErr)
		}
	}()
	deps, err := bootstrap.NewDeps(cfg, db)
	if err != nil {
		fatal(err)
	}
	application := bootstrap.New(cfg, router.New(deps))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := application.Run(ctx); err != nil {
		fatal(err)
	}
}

// fatal 输出错误并以失败状态退出进程。
func fatal(err error) {
	slog.Error("server failed", "error", err)
	os.Exit(1)
}
