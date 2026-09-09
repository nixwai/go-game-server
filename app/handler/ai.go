// Package handler 将 HTTP 请求转换为 AI 模型管理服务调用，并输出统一响应。
package handler

import (
	"github.com/gin-gonic/gin"
	ai "github.com/nixwai/go-game-server/app/dto/ai"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
	"github.com/nixwai/go-game-server/app/service"
)

// AIHandler 是 AI 模型管理相关接口的 HTTP 处理器。
type AIHandler struct {
	// svc 是 AI 管理业务服务。
	svc *service.AIService
	// crypto 负责提供 RSA 公钥。
	crypto *security.CryptoManager
}

// NewAIHandler 创建 AI 管理 HTTP 处理器。
func NewAIHandler(svc *service.AIService, crypto *security.CryptoManager) *AIHandler {
	return &AIHandler{svc: svc, crypto: crypto}
}

// PublicKey 返回 RSA 公钥的 PEM 字符串，供前端加密 API Key 使用。
func (h *AIHandler) PublicKey(c *gin.Context) {
	pem, err := h.crypto.PublicKeyPEM()
	if err != nil {
		response.WriteError(c, response.NewError(response.CodeInternal, "服务器内部错误", err))
		return
	}
	response.Write(c, response.CodeOK, "成功", gin.H{"public_key": pem})
}

// ListProviders 返回当前用户的所有产商配置和模型，列表头部插入只读默认 AI。
func (h *AIHandler) ListProviders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.ListProviders(c.Request.Context(), userID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	// 构建用户产商响应列表。
	result := make([]ai.ProviderResponse, 0, len(list)+1)
	for _, item := range list {
		models := make([]ai.ModelResponse, 0, len(item.Models))
		for _, m := range item.Models {
			models = append(models, ai.NewModelResponse(m, false))
		}
		result = append(result, ai.NewProviderResponse(item.Provider, models, false, item.Provider.APIKeyEncrypted != ""))
	}
	// 在列表头部插入配置文件提供的只读默认 AI。
	defaultAI := h.svc.DefaultAI()
	defaultProvider := model.AIProvider{
		ID:           model.DefaultAIProviderID,
		ProviderName: defaultAI.ProviderName,
		BaseURL:      defaultAI.BaseURL,
		Status:       model.StatusActive,
	}
	defaultModel := model.AIModel{
		ID:        model.DefaultAIModelID,
		ModelName: defaultAI.ModelName,
		Status:    model.StatusActive,
	}
	defaultResp := ai.NewProviderResponse(defaultProvider, []ai.ModelResponse{ai.NewModelResponse(defaultModel, true)}, true, defaultAI.APIKey != "")
	result = append([]ai.ProviderResponse{defaultResp}, result...)
	response.Write(c, response.CodeOK, "成功", result)
}

// CreateProvider 处理新增 AI 产商请求。
func (h *AIHandler) CreateProvider(c *gin.Context, req ai.CreateProviderRequest) {
	userID := middleware.GetUserID(c)
	provider, err := h.svc.CreateProvider(c.Request.Context(), userID, req.ProviderName, req.BaseURL, req.EncryptedAPIKey)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", ai.NewProviderResponse(provider, nil, false, provider.APIKeyEncrypted != ""))
}

// UpdateProvider 处理更新 AI 产商请求。
func (h *AIHandler) UpdateProvider(c *gin.Context, req ai.UpdateProviderRequest) {
	userID := middleware.GetUserID(c)
	params := service.UpdateProviderParams{
		ProviderName:    req.ProviderName,
		BaseURL:         req.BaseURL,
		EncryptedAPIKey: req.EncryptedAPIKey,
		Status:          req.Status,
	}
	provider, err := h.svc.UpdateProvider(c.Request.Context(), userID, req.ID, params)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", ai.NewProviderResponse(provider, nil, false, provider.APIKeyEncrypted != ""))
}

// DeleteProvider 处理删除 AI 产商请求。
func (h *AIHandler) DeleteProvider(c *gin.Context, req ai.DeleteProviderRequest) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteProvider(c.Request.Context(), userID, req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", nil)
}

// CreateModel 处理新增 AI 模型请求。
func (h *AIHandler) CreateModel(c *gin.Context, req ai.CreateModelRequest) {
	userID := middleware.GetUserID(c)
	mdl, err := h.svc.CreateModel(c.Request.Context(), userID, req.ProviderID, req.ModelName)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", ai.NewModelResponse(mdl, false))
}

// UpdateModel 处理更新 AI 模型请求。
func (h *AIHandler) UpdateModel(c *gin.Context, req ai.UpdateModelRequest) {
	userID := middleware.GetUserID(c)
	params := service.UpdateModelParams{ModelName: req.ModelName, Status: req.Status}
	mdl, err := h.svc.UpdateModel(c.Request.Context(), userID, req.ID, params)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", ai.NewModelResponse(mdl, false))
}

// DeleteModel 处理删除 AI 模型请求。
func (h *AIHandler) DeleteModel(c *gin.Context, req ai.DeleteModelRequest) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteModel(c.Request.Context(), userID, req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Write(c, response.CodeOK, "成功", nil)
}
