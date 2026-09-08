// Package database 提供数据库连接初始化能力。
package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// OpenMySQL 根据 DSN 创建 GORM MySQL 连接。
// 禁用 CreateClause 外键约束生成，确保表间关联关系由应用层维护。
func OpenMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
}
