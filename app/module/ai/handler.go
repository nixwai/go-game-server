package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/module/ai/dto"
)

// Handler 是 AI 模型管理相关接口的 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建 AI 管理 HTTP 处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListProviders 返回当前用户的所有产商配置和模型，列表头部插入只读默认 AI。
func (h *Handler) ListProviders(c *gin.Context) ([]dto.ProviderResponse, error) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.ListProviders(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProviderResponse, 0, len(list)+1)
	for _, item := range list {
		models := make([]dto.ModelResponse, 0, len(item.Models))
		for _, m := range item.Models {
			models = append(models, dto.NewModelResponse(m, false))
		}
		result = append(result, dto.NewProviderResponse(item.Provider, models, false, item.Provider.APIKeyEncrypted != ""))
	}
	defaultAI := h.svc.DefaultAI()
	defaultProvider := model.AIProvider{
		ID:           model.DefaultAIProviderID,
		ProviderName: defaultAI.ProviderName,
		BaseURL:      "***********",
		Status:       model.StatusActive,
	}
	defaultModel := model.AIModel{
		ID:        model.DefaultAIModelID,
		ModelName: defaultAI.ModelName,
		Status:    model.StatusActive,
	}
	defaultResp := dto.NewProviderResponse(defaultProvider, []dto.ModelResponse{dto.NewModelResponse(defaultModel, true)}, true, defaultAI.APIKey != "")
	return append([]dto.ProviderResponse{defaultResp}, result...), nil
}

// CreateProvider 处理新增 AI 产商请求，返回创建后的产商信息。
func (h *Handler) CreateProvider(c *gin.Context, req dto.CreateProviderRequest) (dto.ProviderResponse, error) {
	userID := middleware.GetUserID(c)
	provider, err := h.svc.CreateProvider(c.Request.Context(), userID, req.ProviderName, req.BaseURL, req.EncryptedAPIKey)
	if err != nil {
		return dto.ProviderResponse{}, err
	}
	return dto.NewProviderResponse(provider, nil, false, provider.APIKeyEncrypted != ""), nil
}

// UpdateProvider 处理更新 AI 产商请求，返回更新后的产商信息。
func (h *Handler) UpdateProvider(c *gin.Context, req dto.UpdateProviderRequest) (dto.ProviderResponse, error) {
	userID := middleware.GetUserID(c)
	params := UpdateProviderParams{
		ProviderName:    req.ProviderName,
		BaseURL:         req.BaseURL,
		EncryptedAPIKey: req.EncryptedAPIKey,
		Status:          req.Status,
	}
	provider, err := h.svc.UpdateProvider(c.Request.Context(), userID, req.ID, params)
	if err != nil {
		return dto.ProviderResponse{}, err
	}
	return dto.NewProviderResponse(provider, nil, false, provider.APIKeyEncrypted != ""), nil
}

// DeleteProvider 处理删除 AI 产商请求。
func (h *Handler) DeleteProvider(c *gin.Context, req dto.DeleteProviderRequest) (gin.H, error) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteProvider(c.Request.Context(), userID, req.ID); err != nil {
		return nil, err
	}
	return gin.H{}, nil
}

// CreateModel 处理新增 AI 模型请求，返回创建后的模型信息。
func (h *Handler) CreateModel(c *gin.Context, req dto.CreateModelRequest) (dto.ModelResponse, error) {
	userID := middleware.GetUserID(c)
	mdl, err := h.svc.CreateModel(c.Request.Context(), userID, req.ProviderID, req.ModelName)
	if err != nil {
		return dto.ModelResponse{}, err
	}
	return dto.NewModelResponse(mdl, false), nil
}

// UpdateModel 处理更新 AI 模型请求，返回更新后的模型信息。
func (h *Handler) UpdateModel(c *gin.Context, req dto.UpdateModelRequest) (dto.ModelResponse, error) {
	userID := middleware.GetUserID(c)
	params := UpdateModelParams{ModelName: req.ModelName, Status: req.Status}
	mdl, err := h.svc.UpdateModel(c.Request.Context(), userID, req.ID, params)
	if err != nil {
		return dto.ModelResponse{}, err
	}
	return dto.NewModelResponse(mdl, false), nil
}

// DeleteModel 处理删除 AI 模型请求。
func (h *Handler) DeleteModel(c *gin.Context, req dto.DeleteModelRequest) (gin.H, error) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteModel(c.Request.Context(), userID, req.ID); err != nil {
		return nil, err
	}
	return gin.H{}, nil
}
