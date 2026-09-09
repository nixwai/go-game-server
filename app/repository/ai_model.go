package repository

import (
	"context"
	"errors"

	"github.com/nixwai/go-game-server/app/model"
	"gorm.io/gorm"
)

// AIModelRepository 定义 AI 模型配置的数据访问边界。
type AIModelRepository interface {
	// FindByProviderID 查询指定产商下的所有模型配置。
	FindByProviderID(ctx context.Context, providerID uint64) ([]model.AIModel, error)
	// FindByID 根据主键查询单个模型配置。
	FindByID(ctx context.Context, id uint64) (model.AIModel, error)
	// Create 创建一条模型配置记录。
	Create(ctx context.Context, mdl *model.AIModel) error
	// Update 使用 Save 保存模型配置的所有字段。
	Update(ctx context.Context, mdl *model.AIModel) error
	// Delete 删除模型配置记录。
	Delete(ctx context.Context, id uint64) error
	// DeleteByProviderID 删除指定产商下的所有模型配置，用于产商删除时的级联清理。
	DeleteByProviderID(ctx context.Context, providerID uint64) error
}

// GormAIModelRepository 是基于 GORM 的 AIModelRepository 实现。
type GormAIModelRepository struct {
	db *gorm.DB
}

// NewGormAIModelRepository 创建 GORM AI 模型仓储。
func NewGormAIModelRepository(db *gorm.DB) *GormAIModelRepository {
	return &GormAIModelRepository{db: db}
}

// FindByProviderID 查询指定产商下的所有模型配置，按创建时间降序排列。
func (r *GormAIModelRepository) FindByProviderID(ctx context.Context, providerID uint64) ([]model.AIModel, error) {
	var models []model.AIModel
	err := r.db.WithContext(ctx).Where("provider_id = ?", providerID).Order("created_at DESC").Find(&models).Error
	return models, err
}

// FindByID 根据主键查询模型配置，未找到时返回 ErrNotFound。
func (r *GormAIModelRepository) FindByID(ctx context.Context, id uint64) (model.AIModel, error) {
	var mdl model.AIModel
	err := r.db.WithContext(ctx).First(&mdl, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AIModel{}, ErrNotFound
	}
	return mdl, err
}

// Create 创建模型配置记录。
func (r *GormAIModelRepository) Create(ctx context.Context, mdl *model.AIModel) error {
	return r.db.WithContext(ctx).Create(mdl).Error
}

// Update 使用 Save 保存模型配置的所有字段。
func (r *GormAIModelRepository) Update(ctx context.Context, mdl *model.AIModel) error {
	return r.db.WithContext(ctx).Save(mdl).Error
}

// Delete 删除模型配置记录。
func (r *GormAIModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AIModel{}, id).Error
}

// DeleteByProviderID 删除指定产商下的所有模型配置。
func (r *GormAIModelRepository) DeleteByProviderID(ctx context.Context, providerID uint64) error {
	return r.db.WithContext(ctx).Where("provider_id = ?", providerID).Delete(&model.AIModel{}).Error
}
