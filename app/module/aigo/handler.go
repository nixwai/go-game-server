package aigo

import (
	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/module/aigo/dto"
)

// Handler 是围棋对弈相关接口的 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 创建围棋对弈 HTTP 处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GetSetting 返回当前用户的对弈配置。
func (h *Handler) GetSetting(c *gin.Context) (dto.GameSettingResponse, error) {
	userID := middleware.GetUserID(c)
	return h.svc.GetSetting(c.Request.Context(), userID)
}

// UpdateSetting 处理更新对弈配置请求，返回更新后的配置信息。
func (h *Handler) UpdateSetting(c *gin.Context, req dto.UpdateGameSettingRequest) (dto.GameSettingResponse, error) {
	userID := middleware.GetUserID(c)
	params := UpdateSettingParams{
		ActiveModelID:  req.ActiveModelID,
		AllowAIEndGame: req.AllowAIEndGame,
	}
	return h.svc.UpdateSetting(c.Request.Context(), userID, params)
}

// Analyze 处理棋局分析请求，返回 AI 下一步落子或结束申请。
func (h *Handler) Analyze(c *gin.Context, req dto.AnalyzeRequest) (dto.AnalyzeResponse, error) {
	userID := middleware.GetUserID(c)
	return h.svc.Analyze(c.Request.Context(), userID, req)
}
