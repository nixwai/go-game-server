// Command admin-init 交互式创建首个管理员用户。
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/database"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/security"
	"golang.org/x/term"
)

// main 读取管理员凭据、生成 Argon2id 哈希并写入数据库。
func main() {
	cfg, err := config.LoadForAdmin()
	if err != nil {
		fail(err)
	}
	db, err := database.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		fail(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fail(err)
	}
	defer sqlDB.Close()

	reader := bufio.NewReader(os.Stdin)
	username := prompt(reader, "管理员用户名: ")
	if username == "" {
		fail(fmt.Errorf("username is required"))
	}
	fmt.Print("管理员密码: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fail(err)
	}

	password := string(passwordBytes)
	hasher := security.PasswordHasher{Time: cfg.Argon2Time, Memory: cfg.Argon2Memory, Threads: cfg.Argon2Threads, KeyLen: cfg.Argon2KeyLen, SaltLen: cfg.Argon2SaltLen}
	hash, err := hasher.Hash(password)
	if err != nil {
		fail(err)
	}

	repo := repository.NewGormUserRepository(db)
	user := model.User{Username: strings.TrimSpace(username), PasswordHash: hash, Role: model.RoleAdmin, Status: model.StatusActive}
	if err := repo.Create(context.Background(), &user); err != nil {
		fail(err)
	}
	fmt.Printf("管理员创建成功，ID=%d\n", user.ID)
}

// prompt 从终端读取一行文本并去除首尾空白。
func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

// fail 输出初始化错误并以失败状态退出进程。
func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
