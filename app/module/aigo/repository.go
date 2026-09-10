package aigo

import (
	"context"
	"errors"

	"github.com/nixwai/go-game-server/app/model"
	"gorm.io/gorm"
)

// GameSettingRepository 定义围棋对弈配置的数据访问边界。
type GameSettingRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (model.GoGameSetting, error)
	Create(ctx context.Context, setting *model.GoGameSetting) error
	Update(ctx context.Context, setting *model.GoGameSetting) error
}

// GormGameSettingRepository 是基于 GORM 的 GameSettingRepository 实现。
type GormGameSettingRepository struct {
	db *gorm.DB
}

// NewGormGameSettingRepository 创建 GORM 围棋对弈配置仓储。
func NewGormGameSettingRepository(db *gorm.DB) *GormGameSettingRepository {
	return &GormGameSettingRepository{db: db}
}

func (r *GormGameSettingRepository) FindByUserID(ctx context.Context, userID uint64) (model.GoGameSetting, error) {
	var setting model.GoGameSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GoGameSetting{}, model.ErrNotFound
	}
	return setting, err
}

func (r *GormGameSettingRepository) Create(ctx context.Context, setting *model.GoGameSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *GormGameSettingRepository) Update(ctx context.Context, setting *model.GoGameSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}
