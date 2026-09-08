// Package bootstrap 负责构建并运行 HTTP 服务应用。
package bootstrap

import (
	"database/sql"
	"fmt"

	"github.com/nixwai/go-game-server/app/database"
	"gorm.io/gorm"
)

// DB 持有 GORM 连接及其底层标准库连接，用于统一管理数据库生命周期。
type DB struct {
	GORM *gorm.DB
	SQL  *sql.DB
}

// OpenDB 根据 DSN 创建 MySQL 连接并返回可用于关闭的底层句柄。
func OpenDB(dsn string) (*DB, error) {
	db, err := database.OpenMySQL(dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database: %w", err)
	}
	return &DB{GORM: db, SQL: sqlDB}, nil
}

// Close 关闭底层数据库连接。
func (d *DB) Close() error {
	if d == nil || d.SQL == nil {
		return nil
	}
	if err := d.SQL.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
