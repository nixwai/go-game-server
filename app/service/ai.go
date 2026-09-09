// Package service 实现 AI 模型管理相关业务规则。
package service

import (
	"context"
	"strings"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// AIService 编排 AI 产商和模型配置的增删改查，以及 API Key 的加解密流程。
type AIService struct {
	// providers 是 AI 产商数据访问接口。
	providers repository.AIProviderRepository
	// models 是 AI 模型数据访问接口。
	models repository.AIModelRepository
	// crypto 负责 API Key 的加解密。
	crypto *security.CryptoManager
	// defaultAI 是配置文件提供的默认 AI 产商和模型。
	defaultAI config.DefaultAIConfig
}

// ProviderWithModels 是 Service 层返回的产商及其关联模型组合，供 Handler 转换为 DTO。
type ProviderWithModels struct {
	// Provider 是持久化产商实体。
	Provider model.AIProvider
	// Models 是该产商下的所有模型实体。
	Models []model.AIModel
}

// NewAIService 创建 AI 管理服务，并注入外部依赖。
func NewAIService(providers repository.AIProviderRepository, models repository.AIModelRepository, crypto *security.CryptoManager, defaultAI config.DefaultAIConfig) *AIService {
	return &AIService{providers: providers, models: models, crypto: crypto, defaultAI: defaultAI}
}

// DefaultAI 返回配置文件提供的只读默认 AI 产商和模型信息。
func (s *AIService) DefaultAI() config.DefaultAIConfig { return s.defaultAI }

// ListProviders 查询当前用户的所有产商配置和模型，不含默认 AI（由 Handler 层组装）。
func (s *AIService) ListProviders(ctx context.Context, userID uint64) ([]ProviderWithModels, error) {
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
func (s *AIService) CreateProvider(ctx context.Context, userID uint64, providerName, baseURL, encryptedAPIKey string) (model.AIProvider, error) {
	if err := validateProviderName(providerName); err != nil {
		return model.AIProvider{}, err
	}
	if err := validateBaseURL(baseURL); err != nil {
		return model.AIProvider{}, err
	}
	// RSA 解密传输层密文，得到明文 API Key。
	plainKey, err := s.crypto.DecryptTransport(encryptedAPIKey)
	if err != nil {
		return model.AIProvider{}, response.NewError(response.CodeDecryptFailed, "API Key 解密失败", err)
	}
	// AES-256-GCM 加密后入库。
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
func (s *AIService) UpdateProvider(ctx context.Context, userID, providerID uint64, req UpdateProviderParams) (model.AIProvider, error) {
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
		if err := validateStatus(*req.Status); err != nil {
			return model.AIProvider{}, err
		}
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

// DeleteProvider 校验归属权，级联删除产商和关联模型；默认产商不可删除。
func (s *AIService) DeleteProvider(ctx context.Context, userID, providerID uint64) error {
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
	if err := s.models.DeleteByProviderID(ctx, providerID); err != nil {
		return response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	if err := s.providers.Delete(ctx, providerID); err != nil {
		return response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return nil
}

// CreateModel 校验产商归属权，创建模型配置。
func (s *AIService) CreateModel(ctx context.Context, userID, providerID uint64, modelName string) (model.AIModel, error) {
	if err := validateModelName(modelName); err != nil {
		return model.AIModel{}, err
	}
	if err := s.assertProviderOwned(ctx, userID, providerID); err != nil {
		return model.AIModel{}, err
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
func (s *AIService) UpdateModel(ctx context.Context, userID, modelID uint64, req UpdateModelParams) (model.AIModel, error) {
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
		if err := validateStatus(*req.Status); err != nil {
			return model.AIModel{}, err
		}
		mdl.Status = *req.Status
	}
	if err := s.models.Update(ctx, &mdl); err != nil {
		return model.AIModel{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return mdl, nil
}

// DeleteModel 校验归属权，删除模型；默认模型不可删除。
func (s *AIService) DeleteModel(ctx context.Context, userID, modelID uint64) error {
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
func (s *AIService) assertProviderOwned(ctx context.Context, userID, providerID uint64) error {
	provider, err := s.providers.FindByID(ctx, providerID)
	if err != nil {
		return response.NewError(response.CodeNotFound, "产商不存在", err)
	}
	if provider.UserID != userID {
		return response.NewError(response.CodeNotFound, "产商不存在", nil)
	}
	return nil
}

// validateProviderName 校验产商名称长度。
func validateProviderName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 128 {
		return response.NewError(response.CodeValidation, "产商名称长度必须在 1-128 字符之间", nil)
	}
	return nil
}

// validateBaseURL 校验 Base URL 长度。
func validateBaseURL(url string) error {
	url = strings.TrimSpace(url)
	if len(url) < 1 || len(url) > 512 {
		return response.NewError(response.CodeValidation, "Base URL 长度必须在 1-512 字符之间", nil)
	}
	return nil
}

// validateModelName 校验模型名称长度。
func validateModelName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 128 {
		return response.NewError(response.CodeValidation, "模型名称长度必须在 1-128 字符之间", nil)
	}
	return nil
}

// validateStatus 校验状态值是否合法。
func validateStatus(status string) error {
	if status != model.StatusActive && status != model.StatusDisabled {
		return response.NewError(response.CodeValidation, "状态值无效", nil)
	}
	return nil
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
