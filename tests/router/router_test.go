package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type testRepo struct {
	users map[string]model.User
	next  uint64
}

func (r *testRepo) FindByUsername(_ context.Context, name string) (model.User, error) {
	u, ok := r.users[name]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return u, nil
}
func (r *testRepo) FindByID(_ context.Context, id uint64) (model.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return model.User{}, model.ErrNotFound
}
func (r *testRepo) Create(_ context.Context, u *model.User) error {
	if _, ok := r.users[u.Username]; ok {
		return model.ErrDuplicate
	}
	u.ID = r.next
	r.next++
	r.users[u.Username] = *u
	return nil
}

func newRouterForTest() (*gin.Engine, *security.TokenManager) {
	repo := &testRepo{users: map[string]model.User{}, next: 1}
	hasher := security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
	tokens := security.NewTokenManager("01234567890123456789012345678901", "test", time.Hour)
	authSvc := auth.NewService(repo, hasher, tokens)
	authH := auth.NewHandler(authSvc)
	crypto, _ := security.NewCryptoManager("MDEyMzQ1Njc4OWFiY2RlZmdoMTIzNDU2Nzg5YWJjZGVmZ2g=")
	aiSvc := ai.NewService(nil, nil, crypto, config.DefaultAIConfig{ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", ModelName: "gpt-4o-mini", APIKey: "sk-test"})
	aiH := ai.NewHandler(aiSvc, crypto)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID())
	r.GET("/health", ginext.Wrap(func(c *gin.Context) (gin.H, error) { return gin.H{"status": "ok"}, nil }))
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")
	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authH, tokens)
	ai.RegisterRoutes(api, aiH, tokens)
	return r, tokens
}

func request(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// assertCode 从响应体解析业务码并断言。
func assertCode(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()
	if w.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != expectedCode {
		t.Fatalf("expected code %d, got %d, body=%s", expectedCode, body.Code, w.Body.String())
	}
}

func TestAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, tokens := newRouterForTest()
	w := request(r, http.MethodGet, "/health", "", "")
	assertCode(t, w, response.CodeOK)

	w = request(r, http.MethodPost, "/api/v1/auth/register", `{"username":"alice","password":"SecurePass123"}`, "")
	assertCode(t, w, response.CodeOK)
	w = request(r, http.MethodPost, "/api/v1/auth/login", `{"username":"alice","password":"SecurePass123"}`, "")
	assertCode(t, w, response.CodeOK)
	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Token == "" {
		t.Fatal("token missing")
	}
	w = request(r, http.MethodGet, "/api/v1/admin/ping", "", payload.Data.Token)
	assertCode(t, w, response.CodeForbidden)

	adminToken, err := tokens.Generate(model.User{ID: 99, Username: "admin", Role: model.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	w = request(r, http.MethodGet, "/api/v1/admin/ping", "", adminToken)
	assertCode(t, w, response.CodeOK)
	w = request(r, http.MethodGet, "/api/v1/auth/me", "", "bad-token")
	assertCode(t, w, response.CodeTokenInvalid)
}
