// Command migrate 执行版本化数据库迁移。
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nixwai/go-game-server/app/config"
)

// main 根据命令参数执行一次升级或回滚操作。
func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate [up|down]")
		os.Exit(2)
	}

	if err := config.LoadDotEnv(); err != nil {
		fail(err)
	}

	// 迁移只需要数据库 URL，因此不强制要求服务运行所需的 JWT_SECRET。
	databaseURL := os.Getenv("MYSQL_MIGRATE_URL")
	if databaseURL == "" {
		fail(fmt.Errorf("MYSQL_MIGRATE_URL is required"))
	}
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		fail(err)
	}
	defer m.Close()

	if os.Args[1] == "up" {
		err = m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
	} else {
		err = m.Steps(-1)
	}
	if err != nil {
		fail(err)
	}
	fmt.Println("migration completed")
}

// fail 输出迁移错误并以失败状态退出进程。
func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
