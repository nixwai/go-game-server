package router_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/ginext"
	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/model"
	aigo "github.com/nixwai/go-game-server/app/module/aigo"
	"github.com/nixwai/go-game-server/app/module/auth"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

type aigoMockLLM struct {
	content string
	err     error
}

func (m *aigoMockLLM) ChatCompletion(_ context.Context, _ aigo.ChatRequest) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.content, nil
}

type aigoRtSettings struct {
	byID     map[uint64]model.GoGameSetting
	byUserID map[uint64]model.GoGameSetting
	next     uint64
}

func newAigoRtSettings() *aigoRtSettings {
	return &aigoRtSettings{byID: map[uint64]model.GoGameSetting{}, byUserID: map[uint64]model.GoGameSetting{}, next: 1}
}

func (m *aigoRtSettings) FindByUserID(_ context.Context, userID uint64) (model.GoGameSetting, error) {
	s, ok := m.byUserID[userID]
	if !ok {
		return model.GoGameSetting{}, model.ErrNotFound
	}
	return s, nil
}

func (m *aigoRtSettings) Create(_ context.Context, s *model.GoGameSetting) error {
	s.ID = m.next
	m.next++
	m.byID[s.ID] = *s
	m.byUserID[s.UserID] = *s
	return nil
}

func (m *aigoRtSettings) Update(_ context.Context, s *model.GoGameSetting) error {
	m.byID[s.ID] = *s
	m.byUserID[s.UserID] = *s
	return nil
}

func newAIGoRouterForTest(mock *aigoMockLLM) (*gin.Engine, *security.TokenManager) {
	gin.SetMode(gin.TestMode)
	repo := &testRepo{users: map[string]model.User{}, next: 1}
	hasher := security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
	tokens := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	authSvc := auth.NewService(repo, hasher, tokens, -1)
	crypto, _ := security.NewCryptoManager(validMasterKeyB64AI())
	authH := auth.NewHandler(authSvc, crypto)
	aigoSvc := aigo.NewService(newAigoRtSettings(), newMemAIProviders(), newMemAIModels(), crypto, mock, config.DefaultAIConfig{
		ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", ModelName: "gpt-4o-mini", APIKey: "sk-test",
	}, 30*time.Second)
	aigoH := aigo.NewHandler(aigoSvc)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", ginext.Wrap(func(c *gin.Context) (gin.H, error) { return gin.H{"status": "ok"}, nil }))
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")
	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authH, tokens)
	aigo.RegisterRoutes(api, aigoH, tokens)
	return r, tokens
}

func TestAigoGetSettingRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, _ := newAIGoRouterForTest(mock)
	w := request(r, http.MethodGet, "/api/v1/ai/go/setting", "", "")
	assertCode(t, w, response.CodeTokenInvalid)
}

func TestAigoGetSettingSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	w := request(r, http.MethodGet, "/api/v1/ai/go/setting", "", token)
	assertCode(t, w, response.CodeOK)
}

func TestAigoUpdateSettingSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	body := `{"allow_ai_end_game":false}`
	w := request(r, http.MethodPost, "/api/v1/ai/go/setting/update", body, token)
	assertCode(t, w, response.CodeOK)
}

func TestAigoAnalyzeMove(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{content: `{"action":"move","vertex":[1,1]}`}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	body := `{"size":3,"layout":[[0,0,0],[0,0,0],[0,0,0]],"player":1}`
	w := request(r, http.MethodPost, "/api/v1/ai/go/analyze", body, token)
	assertCode(t, w, response.CodeOK)
}

func TestAigoAnalyzeEndGame(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{content: `{"action":"end_game"}`}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	body := `{"size":3,"layout":[[0,0,0],[0,0,0],[0,0,0]],"player":1}`
	w := request(r, http.MethodPost, "/api/v1/ai/go/analyze", body, token)
	assertCode(t, w, response.CodeOK)
}

func TestAigoAnalyzeRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, _ := newAIGoRouterForTest(mock)
	body := `{"size":3,"layout":[[0,0,0],[0,0,0],[0,0,0]],"player":1}`
	w := request(r, http.MethodPost, "/api/v1/ai/go/analyze", body, "")
	assertCode(t, w, response.CodeTokenInvalid)
}

func TestAigoAnalyzeMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	w := request(r, http.MethodPost, "/api/v1/ai/go/analyze", `{}`, token)
	assertCode(t, w, response.CodeValidation)
}

func TestAigoAnalyzeInvalidLayout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &aigoMockLLM{}
	r, tokens := newAIGoRouterForTest(mock)
	token := generateAITestToken(t, tokens, 1)
	body := `{"size":3,"layout":[[0,0]],"player":1}`
	w := request(r, http.MethodPost, "/api/v1/ai/go/analyze", body, token)
	assertCode(t, w, response.CodeValidation)
}
