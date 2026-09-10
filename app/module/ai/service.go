package ai

import (
	"context"
	"strings"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// Service 编排 AI 产商和模型配置的增删改查，以及 API Key 的加解密流程。
type Service struct {
	providers    ProviderRepository
	models       ModelRepository
	crypto       *security.CryptoManager
	defaultAI    config.DefaultAIConfig
	maxProviders int
	maxModels    int
}

// ProviderWithModels 是 Service 层返回的产商及其关联模型组合，供 Handler 转换为 DTO。
type ProviderWithModels struct {
	Provider model.AIProvider
	Models   []model.AIModel
}

// UpdateProviderParams 是更新产商的参数。
type UpdateProviderParams struct {
	ProviderName    *string
	BaseURL         *string
	EncryptedAPIKey *string
	Status          *string
}

// UpdateModelParams 是更新模型的参数。
type UpdateModelParams struct {
	ModelName *string
	Status    *string
}

// NewService 创建 AI 管理服务，并注入外部依赖。
func NewService(providers ProviderRepository, models ModelRepository, crypto *security.CryptoManager, cfg config.Config) *Service {
	return &Service{providers: providers, models: models, crypto: crypto, defaultAI: cfg.DefaultAI, maxProviders: cfg.MaxProvidersPerUser, maxModels: cfg.MaxModelsPerProvider}
}

// DefaultAI 返回配置文件提供的只读默认 AI 产商和模型信息。
func (s *Service) DefaultAI() config.DefaultAIConfig { return s.defaultAI }

// ListProviders 查询当前用户的所有产商配置和模型，不含默认 AI（由 Handler 层组装）。
func (s *Service) ListProviders(ctx context.Context, userID uint64) ([]ProviderWithModels, error) {
	providers, err := s.providers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	result := make([]ProviderWithModels, 0, len(providers))
	for _, p := range providers {
		models, err := s.models.FindByProviderID(ctx, p.ID)
		if err != nil {
			return nil, response.NewError(response.CodeInternal, "服务器内部错误", err)
		}
		result = append(result, ProviderWithModels{Provider: p, Models: models})
	}
	return result, nil
}

// CreateProvider 校验输入、解密传输层 API Key、AES 加密后入库。
func (s *Service) CreateProvider(ctx context.Context, userID uint64, providerName, baseURL, encryptedAPIKey string) (model.AIProvider, error) {
	if s.maxProviders >= 0 {
		count, err := s.providers.CountByUserID(ctx, userID)
		if err != nil {
			return model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
		}
		if count >= int64(s.maxProviders) {
			return model.AIProvider{}, response.NewError(response.CodeLimitExceeded, "产商数量已达上限", nil)
		}
	}
	if err := validateProviderName(providerName); err != nil {
		return model.AIProvider{}, err
	}
	if err := validateBaseURL(baseURL); err != nil {
		return model.AIProvider{}, err
	}
	plainKey, err := s.crypto.DecryptTransport(encryptedAPIKey)
	if err != nil {
		return model.AIProvider{}, response.NewError(response.CodeDecryptFailed, "API Key 解密失败", err)
	}
	storageCipher, err := s.crypto.EncryptForStorage(plainKey)
	if err != nil {
		return model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	provider := model.AIProvider{
		UserID:          userID,
		ProviderName:    strings.TrimSpace(providerName),
		BaseURL:         strings.TrimSpace(baseURL),
		APIKeyEncrypted: storageCipher,
		Status:          model.StatusActive,
	}
	if err := s.providers.Create(ctx, &provider); err != nil {
		return model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return provider, nil
}

// UpdateProvider 校验归属权，更新产商配置；如提供新 encrypted_api_key 则重新加解密。
func (s *Service) UpdateProvider(ctx context.Context, userID, providerID uint64, req UpdateProviderParams) (model.AIProvider, error) {
	if providerID == model.DefaultAIProviderID {
		return model.AIProvider{}, response.NewError(response.CodeDefaultAIReadOnly, "默认 AI 配置不可修改", nil)
	}
	provider, err := s.providers.FindByID(ctx, providerID)
	if err != nil {
		return model.AIProvider{}, response.NewError(response.CodeNotFound, "产商不存在", err)
	}
	if provider.UserID != userID {
		return model.AIProvider{}, response.NewError(response.CodeNotFound, "产商不存在", nil)
	}
	if req.ProviderName != nil {
		if err := validateProviderName(*req.ProviderName); err != nil {
			return model.AIProvider{}, err
		}
		provider.ProviderName = strings.TrimSpace(*req.ProviderName)
	}
	if req.BaseURL != nil {
		if err := validateBaseURL(*req.BaseURL); err != nil {
			return model.AIProvider{}, err
		}
		provider.BaseURL = strings.TrimSpace(*req.BaseURL)
	}
	if req.Status != nil {
		provider.Status = *req.Status
	}
	if req.EncryptedAPIKey != nil && *req.EncryptedAPIKey != "" {
		plainKey, err := s.crypto.DecryptTransport(*req.EncryptedAPIKey)
		if err != nil {
			return model.AIProvider{}, response.NewError(response.CodeDecryptFailed, "API Key 解密失败", err)
		}
		storageCipher, err := s.crypto.EncryptForStorage(plainKey)
		if err != nil {
			return model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
		}
		provider.APIKeyEncrypted = storageCipher
	}
	if err := s.providers.Update(ctx, &provider); err != nil {
		return model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return provider, nil
}

// DeleteProvider 校验归属权，在单个事务中删除产商及其全部模型。
func (s *Service) DeleteProvider(ctx context.Context, userID, providerID uint64) error {
	if providerID == model.DefaultAIProviderID {
		return response.NewError(response.CodeDefaultAIReadOnly, "默认 AI 配置不可删除", nil)
	}
	provider, err := s.providers.FindByID(ctx, providerID)
	if err != nil {
		return response.NewError(response.CodeNotFound, "产商不存在", err)
	}
	if provider.UserID != userID {
		return response.NewError(response.CodeNotFound, "产商不存在", nil)
	}
	if err := s.providers.DeleteWithModels(ctx, providerID); err != nil {
		return response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return nil
}

// CreateModel 校验产商归属权，创建模型配置。
func (s *Service) CreateModel(ctx context.Context, userID, providerID uint64, modelName string) (model.AIModel, error) {
	if err := validateModelName(modelName); err != nil {
		return model.AIModel{}, err
	}
	if err := s.assertProviderOwned(ctx, userID, providerID); err != nil {
		return model.AIModel{}, err
	}
	if s.maxModels >= 0 {
		count, err := s.models.CountByProviderID(ctx, providerID)
		if err != nil {
			return model.AIModel{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
		}
		if count >= int64(s.maxModels) {
			return model.AIModel{}, response.NewError(response.CodeLimitExceeded, "模型数量已达上限", nil)
		}
	}
	mdl := model.AIModel{
		ProviderID: providerID,
		ModelName:  strings.TrimSpace(modelName),
		Status:     model.StatusActive,
	}
	if err := s.models.Create(ctx, &mdl); err != nil {
		return model.AIModel{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return mdl, nil
}

// UpdateModel 校验模型归属权，更新模型配置。
func (s *Service) UpdateModel(ctx context.Context, userID, modelID uint64, req UpdateModelParams) (model.AIModel, error) {
	if modelID == model.DefaultAIModelID {
		return model.AIModel{}, response.NewError(response.CodeDefaultAIReadOnly, "默认模型不可修改", nil)
	}
	mdl, err := s.models.FindByID(ctx, modelID)
	if err != nil {
		return model.AIModel{}, response.NewError(response.CodeNotFound, "模型不存在", err)
	}
	if err := s.assertProviderOwned(ctx, userID, mdl.ProviderID); err != nil {
		return model.AIModel{}, err
	}
	if req.ModelName != nil {
		if err := validateModelName(*req.ModelName); err != nil {
			return model.AIModel{}, err
		}
		mdl.ModelName = strings.TrimSpace(*req.ModelName)
	}
	if req.Status != nil {
		mdl.Status = *req.Status
	}
	if err := s.models.Update(ctx, &mdl); err != nil {
		return model.AIModel{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return mdl, nil
}

// DeleteModel 校验归属权，删除模型；默认模型不可删除。
func (s *Service) DeleteModel(ctx context.Context, userID, modelID uint64) error {
	if modelID == model.DefaultAIModelID {
		return response.NewError(response.CodeDefaultAIReadOnly, "默认模型不可删除", nil)
	}
	mdl, err := s.models.FindByID(ctx, modelID)
	if err != nil {
		return response.NewError(response.CodeNotFound, "模型不存在", err)
	}
	if err := s.assertProviderOwned(ctx, userID, mdl.ProviderID); err != nil {
		return err
	}
	if err := s.models.Delete(ctx, modelID); err != nil {
		return response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return nil
}

// assertProviderOwned 校验产商存在且属于指定用户。
func (s *Service) assertProviderOwned(ctx context.Context, userID, providerID uint64) error {
	provider, err := s.providers.FindByID(ctx, providerID)
	if err != nil {
		return response.NewError(response.CodeNotFound, "产商不存在", err)
	}
	if provider.UserID != userID {
		return response.NewError(response.CodeNotFound, "产商不存在", nil)
	}
	return nil
}
