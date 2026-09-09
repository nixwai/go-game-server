package ai

import (
	"context"
	"errors"

	"github.com/nixwai/go-game-server/app/model"
	"gorm.io/gorm"
)

// ProviderRepository 定义 AI 产商配置的数据访问边界。
type ProviderRepository interface {
	FindByUserID(ctx context.Context, userID uint64) ([]model.AIProvider, error)
	FindByID(ctx context.Context, id uint64) (model.AIProvider, error)
	Create(ctx context.Context, provider *model.AIProvider) error
	Update(ctx context.Context, provider *model.AIProvider) error
	Delete(ctx context.Context, id uint64) error
}

// ModelRepository 定义 AI 模型配置的数据访问边界。
type ModelRepository interface {
	FindByProviderID(ctx context.Context, providerID uint64) ([]model.AIModel, error)
	FindByID(ctx context.Context, id uint64) (model.AIModel, error)
	Create(ctx context.Context, mdl *model.AIModel) error
	Update(ctx context.Context, mdl *model.AIModel) error
	Delete(ctx context.Context, id uint64) error
	DeleteByProviderID(ctx context.Context, providerID uint64) error
}

// GormProviderRepository 是基于 GORM 的 ProviderRepository 实现。
type GormProviderRepository struct {
	db *gorm.DB
}

// NewGormProviderRepository 创建 GORM AI 产商仓储。
func NewGormProviderRepository(db *gorm.DB) *GormProviderRepository {
	return &GormProviderRepository{db: db}
}

func (r *GormProviderRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.AIProvider, error) {
	var providers []model.AIProvider
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&providers).Error
	return providers, err
}

func (r *GormProviderRepository) FindByID(ctx context.Context, id uint64) (model.AIProvider, error) {
	var provider model.AIProvider
	err := r.db.WithContext(ctx).First(&provider, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AIProvider{}, model.ErrNotFound
	}
	return provider, err
}

func (r *GormProviderRepository) Create(ctx context.Context, provider *model.AIProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *GormProviderRepository) Update(ctx context.Context, provider *model.AIProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

func (r *GormProviderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AIProvider{}, id).Error
}

// GormModelRepository 是基于 GORM 的 ModelRepository 实现。
type GormModelRepository struct {
	db *gorm.DB
}

// NewGormModelRepository 创建 GORM AI 模型仓储。
func NewGormModelRepository(db *gorm.DB) *GormModelRepository {
	return &GormModelRepository{db: db}
}

func (r *GormModelRepository) FindByProviderID(ctx context.Context, providerID uint64) ([]model.AIModel, error) {
	var models []model.AIModel
	err := r.db.WithContext(ctx).Where("provider_id = ?", providerID).Order("created_at DESC").Find(&models).Error
	return models, err
}

func (r *GormModelRepository) FindByID(ctx context.Context, id uint64) (model.AIModel, error) {
	var mdl model.AIModel
	err := r.db.WithContext(ctx).First(&mdl, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AIModel{}, model.ErrNotFound
	}
	return mdl, err
}

func (r *GormModelRepository) Create(ctx context.Context, mdl *model.AIModel) error {
	return r.db.WithContext(ctx).Create(mdl).Error
}

func (r *GormModelRepository) Update(ctx context.Context, mdl *model.AIModel) error {
	return r.db.WithContext(ctx).Save(mdl).Error
}

func (r *GormModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AIModel{}, id).Error
}

func (r *GormModelRepository) DeleteByProviderID(ctx context.Context, providerID uint64) error {
	return r.db.WithContext(ctx).Where("provider_id = ?", providerID).Delete(&model.AIModel{}).Error
}
