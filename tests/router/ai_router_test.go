package router_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/ginext"

	"github.com/nixwai/go-game-server/app/middleware"
	"github.com/nixwai/go-game-server/app/model"
	ai "github.com/nixwai/go-game-server/app/module/ai"
	auth "github.com/nixwai/go-game-server/app/module/auth"

	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// --- 内存仓储 mock（用于 AI 路由测试）---

type memAIProviders struct {
	byID map[uint64]model.AIProvider
	next uint64
}

func newMemAIProviders() *memAIProviders {
	return &memAIProviders{byID: map[uint64]model.AIProvider{}, next: 1}
}
func (m *memAIProviders) FindByUserID(_ context.Context, userID uint64) ([]model.AIProvider, error) {
	var result []model.AIProvider
	for _, p := range m.byID {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *memAIProviders) FindByID(_ context.Context, id uint64) (model.AIProvider, error) {
	p, ok := m.byID[id]
	if !ok {
		return model.AIProvider{}, model.ErrNotFound
	}
	return p, nil
}
func (m *memAIProviders) Create(_ context.Context, p *model.AIProvider) error {
	p.ID = m.next
	m.next++
	m.byID[p.ID] = *p
	return nil
}
func (m *memAIProviders) Update(_ context.Context, p *model.AIProvider) error {
	m.byID[p.ID] = *p
	return nil
}
func (m *memAIProviders) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}
func (m *memAIProviders) DeleteWithModels(_ context.Context, providerID uint64) error {
	delete(m.byID, providerID)
	return nil
}

type memAIModels struct {
	byID map[uint64]model.AIModel
	next uint64
}

func newMemAIModels() *memAIModels {
	return &memAIModels{byID: map[uint64]model.AIModel{}, next: 1}
}
func (m *memAIModels) FindByProviderID(_ context.Context, providerID uint64) ([]model.AIModel, error) {
	var result []model.AIModel
	for _, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			result = append(result, mdl)
		}
	}
	return result, nil
}
func (m *memAIModels) FindByID(_ context.Context, id uint64) (model.AIModel, error) {
	mdl, ok := m.byID[id]
	if !ok {
		return model.AIModel{}, model.ErrNotFound
	}
	return mdl, nil
}
func (m *memAIModels) Create(_ context.Context, mdl *model.AIModel) error {
	mdl.ID = m.next
	m.next++
	m.byID[mdl.ID] = *mdl
	return nil
}
func (m *memAIModels) Update(_ context.Context, mdl *model.AIModel) error {
	m.byID[mdl.ID] = *mdl
	return nil
}
func (m *memAIModels) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}
func (m *memAIModels) DeleteByProviderID(_ context.Context, providerID uint64) error {
	for id, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			delete(m.byID, id)
		}
	}
	return nil
}

func validMasterKeyB64AI() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func newAIRouterForTest() (*gin.Engine, *security.TokenManager, *security.CryptoManager) {
	repo := &testRepo{users: map[string]model.User{}, next: 1}
	hasher := security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
	tokens := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	authSvc := auth.NewService(repo, hasher, tokens, -1)
	crypto, _ := security.NewCryptoManager(validMasterKeyB64AI())
	authH := auth.NewHandler(authSvc, crypto)
	aiSvc := ai.NewService(newMemAIProviders(), newMemAIModels(), crypto, config.DefaultAIConfig{
		ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", ModelName: "gpt-4o-mini", APIKey: "sk-default",
	})
	aiH := ai.NewHandler(aiSvc)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", ginext.Wrap(func(c *gin.Context) (gin.H, error) { return gin.H{"status": "ok"}, nil }))
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")
	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authH, tokens)
	ai.RegisterRoutes(api, aiH, tokens)
	return r, tokens, crypto
}

func TestAIListProvidersIncludesDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	w := request(r, http.MethodGet, "/api/v1/ai/providers/list", "", token)
	assertCode(t, w, response.CodeOK)
	var body struct {
		Data []struct {
			ID           uint64 `json:"id"`
			IsDefault    bool   `json:"is_default"`
			ProviderName string `json:"provider_name"`
			HasAPIKey    bool   `json:"has_api_key"`
			Models       []struct {
				IsDefault bool `json:"is_default"`
			} `json:"models"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 default provider, got %d", len(body.Data))
	}
	if !body.Data[0].IsDefault || body.Data[0].ID != 0 {
		t.Fatal("first provider should be default with ID=0")
	}
	if !body.Data[0].HasAPIKey {
		t.Fatal("default provider should have API key")
	}
	if len(body.Data[0].Models) != 1 || !body.Data[0].Models[0].IsDefault {
		t.Fatal("default provider should have 1 default model")
	}
}

func TestAICreateProviderDecryptFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	body := `{"provider_name":"OpenAI","base_url":"https://api.openai.com/v1","encrypted_api_key":"invalid"}`
	w := request(r, http.MethodPost, "/api/v1/ai/providers/create", body, token)
	assertCode(t, w, response.CodeDecryptFailed)
}

func TestAICreateProviderMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	w := request(r, http.MethodPost, "/api/v1/ai/providers/create", `{}`, token)
	assertCode(t, w, response.CodeValidation)
}

func TestAIUpdateDefaultProviderRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	body := `{"id":0,"provider_name":"NewName"}`
	w := request(r, http.MethodPost, "/api/v1/ai/providers/update", body, token)
	assertCode(t, w, response.CodeDefaultAIReadOnly)
}

func TestAIDeleteDefaultProviderRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	body := `{"id":0}`
	w := request(r, http.MethodPost, "/api/v1/ai/providers/delete", body, token)
	assertCode(t, w, response.CodeDefaultAIReadOnly)
}

func TestAIDeleteDefaultModelRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	body := `{"id":0}`
	w := request(r, http.MethodPost, "/api/v1/ai/models/delete", body, token)
	assertCode(t, w, response.CodeDefaultAIReadOnly)
}

func TestAIUpdateDefaultModelRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens, _ := newAIRouterForTest()
	token := generateAITestToken(t, tokens, 1)
	body := `{"id":0,"model_name":"new-name"}`
	w := request(r, http.MethodPost, "/api/v1/ai/models/update", body, token)
	assertCode(t, w, response.CodeDefaultAIReadOnly)
}

func generateAITestToken(t *testing.T, tokens *security.TokenManager, userID uint64) string {
	t.Helper()
	token, err := tokens.Generate(model.User{ID: userID, Username: "alice", Role: model.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	return token
}
