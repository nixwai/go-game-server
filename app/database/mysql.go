// Package database 提供数据库连接初始化能力。
package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenMySQL 根据 DSN 创建 GORM MySQL 连接。
func OpenMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
