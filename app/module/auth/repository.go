package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/nixwai/go-game-server/app/model"
	"gorm.io/gorm"
)

// UserRepository 定义用户数据访问边界。
type UserRepository interface {
	// FindByUsername 根据登录名查询用户。
	FindByUsername(ctx context.Context, username string) (model.User, error)
	// FindByID 根据用户主键查询用户。
	FindByID(ctx context.Context, id uint64) (model.User, error)
	// Create 创建一条用户记录。
	Create(ctx context.Context, user *model.User) error
	// UpdatePassword 更新指定用户的密码哈希。
	UpdatePassword(ctx context.Context, id uint64, passwordHash string) error
}

// GormUserRepository 是基于 GORM 的 UserRepository 实现。
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository 创建 GORM 用户仓储。
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByUsername 根据用户名查询用户。
func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, model.ErrNotFound
	}
	return user, err
}

// FindByID 根据主键查询用户，并将 GORM 的未找到错误转换为领域错误。
func (r *GormUserRepository) FindByID(ctx context.Context, id uint64) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, model.ErrNotFound
	}
	return user, err
}

// Create 创建用户，并将 MySQL 唯一键冲突统一转换为 ErrDuplicate。
func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil && isDuplicate(err) {
		return model.ErrDuplicate
	}
	return err
}

// UpdatePassword 更新指定用户的密码哈希，未匹配到记录时返回 ErrNotFound。
func (r *GormUserRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

// isDuplicate 识别 MySQL 驱动返回的唯一键冲突错误。
func isDuplicate(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "1062")
}
