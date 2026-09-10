package aigo

import (
	"context"
	"errors"
	"time"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/module/ai"
	"github.com/nixwai/go-game-server/app/module/aigo/dto"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// Service 编排对弈配置管理和棋局分析主流程。
type Service struct {
	settings  GameSettingRepository
	providers ai.ProviderRepository
	models    ai.ModelRepository
	crypto    *security.CryptoManager
	client    LLMClient
	defaultAI config.DefaultAIConfig
	timeout   time.Duration
}

// NewService 创建围棋对弈服务，注入全部外部依赖。
func NewService(
	settings GameSettingRepository,
	providers ai.ProviderRepository,
	models ai.ModelRepository,
	crypto *security.CryptoManager,
	client LLMClient,
	defaultAI config.DefaultAIConfig,
	timeout time.Duration,
) *Service {
	return &Service{
		settings:  settings,
		providers: providers,
		models:    models,
		crypto:    crypto,
		client:    client,
		defaultAI: defaultAI,
		timeout:   timeout,
	}
}

// UpdateSettingParams 是更新对弈配置的参数。
type UpdateSettingParams struct {
	ActiveModelID  *uint64
	AllowAIEndGame *bool
}

// GetSetting 返回当前用户的对弈配置，不存在时自动创建默认配置。
func (s *Service) GetSetting(ctx context.Context, userID uint64) (dto.GameSettingResponse, error) {
	setting, err := s.getOrCreateSetting(ctx, userID)
	if err != nil {
		return dto.GameSettingResponse{}, err
	}
	return s.buildSettingResponse(ctx, setting)
}

// UpdateSetting 校验并更新用户对弈配置。
func (s *Service) UpdateSetting(ctx context.Context, userID uint64, req UpdateSettingParams) (dto.GameSettingResponse, error) {
	setting, err := s.getOrCreateSetting(ctx, userID)
	if err != nil {
		return dto.GameSettingResponse{}, err
	}
	if req.ActiveModelID != nil {
		if err := s.validateModelForUser(ctx, userID, *req.ActiveModelID); err != nil {
			return dto.GameSettingResponse{}, err
		}
		setting.ActiveModelID = *req.ActiveModelID
	}
	if req.AllowAIEndGame != nil {
		setting.AllowAIEndGame = *req.AllowAIEndGame
	}
	if err := s.settings.Update(ctx, &setting); err != nil {
		return dto.GameSettingResponse{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return s.buildSettingResponse(ctx, setting)
}

// Analyze 接收棋局数据，调用 AI 模型分析后返回落子或结束申请。
func (s *Service) Analyze(ctx context.Context, userID uint64, req dto.AnalyzeRequest) (dto.AnalyzeResponse, error) {
	if err := validateSnapshot(req); err != nil {
		return dto.AnalyzeResponse{}, err
	}

	setting, err := s.getOrCreateSetting(ctx, userID)
	if err != nil {
		return dto.AnalyzeResponse{}, err
	}

	baseURL, apiKey, modelName, err := s.resolveAIConfig(ctx, userID, setting.ActiveModelID)
	if err != nil {
		return dto.AnalyzeResponse{}, err
	}

	systemPrompt := buildSystemPrompt(setting.AllowAIEndGame)
	userPrompt, err := buildUserPrompt(req)
	if err != nil {
		return dto.AnalyzeResponse{}, response.NewError(response.CodeValidation, "棋局数据无效", err)
	}

	chatReq := ChatRequest{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Timeout: s.timeout,
	}

	content, err := s.client.ChatCompletion(ctx, chatReq)
	if err != nil {
		return dto.AnalyzeResponse{}, response.NewError(response.CodeAICallFailed, "AI 调用失败", err)
	}

	size, _ := parseSize(req.Size)
	result, err := parseAIResponse(content, size)
	if err != nil {
		return dto.AnalyzeResponse{}, response.NewError(response.CodeAIResponseInvalid, "AI 返回格式无效", err)
	}

	if result.Action == "end_game" && !setting.AllowAIEndGame {
		return dto.AnalyzeResponse{}, response.NewError(response.CodeAIResponseInvalid, "当前配置不允许 AI 申请结束游戏", nil)
	}

	resp := dto.AnalyzeResponse{Action: result.Action}
	if result.Vertex != nil {
		resp.Vertex = result.Vertex
	}
	return resp, nil
}

// getOrCreateSetting 查询用户配置，不存在时创建默认配置。
func (s *Service) getOrCreateSetting(ctx context.Context, userID uint64) (model.GoGameSetting, error) {
	setting, err := s.settings.FindByUserID(ctx, userID)
	if err == nil {
		return setting, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return model.GoGameSetting{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	setting = model.GoGameSetting{
		UserID:         userID,
		ActiveModelID:  model.DefaultAIModelID,
		AllowAIEndGame: true,
	}
	if err := s.settings.Create(ctx, &setting); err != nil {
		return model.GoGameSetting{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return setting, nil
}

// buildSettingResponse 将持久化配置转换为对外响应，补充模型信息。
func (s *Service) buildSettingResponse(ctx context.Context, setting model.GoGameSetting) (dto.GameSettingResponse, error) {
	resp := dto.GameSettingResponse{
		ActiveModelID:  setting.ActiveModelID,
		AllowAIEndGame: setting.AllowAIEndGame,
	}
	if setting.ActiveModelID == model.DefaultAIModelID {
		resp.ModelName = s.defaultAI.ModelName
		resp.IsDefault = true
		resp.HasAPIKey = s.defaultAI.APIKey != ""
		return resp, nil
	}
	mdl, err := s.models.FindByID(ctx, setting.ActiveModelID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return dto.GameSettingResponse{}, response.NewError(response.CodeNotFound, "模型不存在", err)
		}
		return dto.GameSettingResponse{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	resp.ModelName = mdl.ModelName
	resp.IsDefault = false
	provider, err := s.providers.FindByID(ctx, mdl.ProviderID)
	if err != nil {
		return dto.GameSettingResponse{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	resp.HasAPIKey = provider.APIKeyEncrypted != ""
	return resp, nil
}

// validateModelForUser 校验模型存在、属于用户且处于 active 状态。
func (s *Service) validateModelForUser(ctx context.Context, userID, modelID uint64) error {
	if modelID == model.DefaultAIModelID {
		return nil
	}
	_, _, err := s.resolveModelAndProvider(ctx, userID, modelID)
	return err
}

// resolveModelAndProvider 校验模型和产商的存在性、归属权和状态，返回合法的模型和产商。
func (s *Service) resolveModelAndProvider(ctx context.Context, userID, modelID uint64) (model.AIModel, model.AIProvider, error) {
	mdl, err := s.models.FindByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeNotFound, "模型不存在", err)
		}
		return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	if mdl.Status != model.StatusActive {
		return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeModelInactive, "模型已禁用", nil)
	}
	provider, err := s.providers.FindByID(ctx, mdl.ProviderID)
	if err != nil {
		return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeNotFound, "产商不存在", err)
	}
	if provider.UserID != userID {
		return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeNotFound, "模型不存在", nil)
	}
	if provider.Status != model.StatusActive {
		return model.AIModel{}, model.AIProvider{}, response.NewError(response.CodeModelInactive, "产商已禁用", nil)
	}
	return mdl, provider, nil
}

// resolveAIConfig 解析激活模型对应的 API 调用参数。
func (s *Service) resolveAIConfig(ctx context.Context, userID, modelID uint64) (baseURL, apiKey, modelName string, err error) {
	if modelID == model.DefaultAIModelID {
		return s.defaultAI.BaseURL, s.defaultAI.APIKey, s.defaultAI.ModelName, nil
	}
	mdl, provider, err := s.resolveModelAndProvider(ctx, userID, modelID)
	if err != nil {
		return "", "", "", err
	}
	plainKey, err := s.crypto.DecryptFromStorage(provider.APIKeyEncrypted)
	if err != nil {
		return "", "", "", response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return provider.BaseURL, plainKey, mdl.ModelName, nil
}

// validateSnapshot 校验棋局数据的结构完整性。
func validateSnapshot(req dto.AnalyzeRequest) error {
	size, err := parseSize(req.Size)
	if err != nil {
		return response.NewError(response.CodeValidation, "棋盘尺寸无效", err)
	}
	if size < 2 || size > 52 {
		return response.NewError(response.CodeValidation, "棋盘尺寸必须在 2-52 之间", nil)
	}
	if len(req.Layout) != size {
		return response.NewError(response.CodeValidation, "棋盘布局行数与尺寸不匹配", nil)
	}
	for _, row := range req.Layout {
		if len(row) != size {
			return response.NewError(response.CodeValidation, "棋盘布局列数与尺寸不匹配", nil)
		}
		for _, cell := range row {
			if cell != 0 && cell != 1 && cell != -1 {
				return response.NewError(response.CodeValidation, "棋盘包含无效棋子标记", nil)
			}
		}
	}
	if req.Player != 1 && req.Player != -1 {
		return response.NewError(response.CodeValidation, "执棋方必须为 1 或 -1", nil)
	}
	if req.Ko != nil {
		if req.Ko.Sign != 1 && req.Ko.Sign != -1 {
			return response.NewError(response.CodeValidation, "劫子标记必须为 1 或 -1", nil)
		}
	}
	return nil
}
