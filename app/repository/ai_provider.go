package repository

import (
	"context"
	"errors"

	"github.com/nixwai/go-game-server/app/model"
	"gorm.io/gorm"
)

// AIProviderRepository 定义 AI 产商配置的数据访问边界。
type AIProviderRepository interface {
	// FindByUserID 查询指定用户的所有产商配置。
	FindByUserID(ctx context.Context, userID uint64) ([]model.AIProvider, error)
	// FindByID 根据 ID 查询单个产商配置。
	FindByID(ctx context.Context, id uint64) (model.AIProvider, error)
	// Create 创建一条产商配置记录。
	Create(ctx context.Context, provider *model.AIProvider) error
	// Update 更新产商配置的非空字段。
	Update(ctx context.Context, provider *model.AIProvider) error
	// Delete 删除产商配置。
	Delete(ctx context.Context, id uint64) error
}

// GormAIProviderRepository 是基于 GORM 的 AIProviderRepository 实现。
type GormAIProviderRepository struct {
	db *gorm.DB
}

// NewGormAIProviderRepository 创建 GORM AI 产商仓储。
func NewGormAIProviderRepository(db *gorm.DB) *GormAIProviderRepository {
	return &GormAIProviderRepository{db: db}
}

// FindByUserID 使用参数化条件查询指定用户的所有产商配置，按创建时间降序排列。
func (r *GormAIProviderRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.AIProvider, error) {
	var providers []model.AIProvider
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&providers).Error
	return providers, err
}

// FindByID 根据主键查询产商配置，未找到时返回 ErrNotFound。
func (r *GormAIProviderRepository) FindByID(ctx context.Context, id uint64) (model.AIProvider, error) {
	var provider model.AIProvider
	err := r.db.WithContext(ctx).First(&provider, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AIProvider{}, ErrNotFound
	}
	return provider, err
}

// Create 创建产商配置记录。
func (r *GormAIProviderRepository) Create(ctx context.Context, provider *model.AIProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// Update 使用 Save 保存产商配置的所有字段。
func (r *GormAIProviderRepository) Update(ctx context.Context, provider *model.AIProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// Delete 删除产商配置记录。
func (r *GormAIProviderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AIProvider{}, id).Error
}
