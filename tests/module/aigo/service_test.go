package aigo_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/module/ai"
	aigo "github.com/nixwai/go-game-server/app/module/aigo"
	"github.com/nixwai/go-game-server/app/module/aigo/dto"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// --- mock LLM client ---

type mockLLMClient struct {
	content string
	err     error
}

func (m *mockLLMClient) ChatCompletion(_ context.Context, _ aigo.ChatRequest) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.content, nil
}

// --- mock GameSettingRepository ---

type memoryGameSettings struct {
	byID     map[uint64]model.GoGameSetting
	byUserID map[uint64]model.GoGameSetting
	next     uint64
}

func newMemoryGameSettings() *memoryGameSettings {
	return &memoryGameSettings{byID: map[uint64]model.GoGameSetting{}, byUserID: map[uint64]model.GoGameSetting{}, next: 1}
}

func (m *memoryGameSettings) FindByUserID(_ context.Context, userID uint64) (model.GoGameSetting, error) {
	s, ok := m.byUserID[userID]
	if !ok {
		return model.GoGameSetting{}, model.ErrNotFound
	}
	return s, nil
}

func (m *memoryGameSettings) Create(_ context.Context, s *model.GoGameSetting) error {
	s.ID = m.next
	m.next++
	m.byID[s.ID] = *s
	m.byUserID[s.UserID] = *s
	return nil
}

func (m *memoryGameSettings) Update(_ context.Context, s *model.GoGameSetting) error {
	m.byID[s.ID] = *s
	m.byUserID[s.UserID] = *s
	return nil
}

// --- mock ai.ProviderRepository ---

type memProviders struct {
	byID map[uint64]model.AIProvider
	next uint64
}

func newMemProviders() *memProviders {
	return &memProviders{byID: map[uint64]model.AIProvider{}, next: 1}
}

func (m *memProviders) FindByUserID(_ context.Context, userID uint64) ([]model.AIProvider, error) {
	var result []model.AIProvider
	for _, p := range m.byID {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *memProviders) FindByID(_ context.Context, id uint64) (model.AIProvider, error) {
	p, ok := m.byID[id]
	if !ok {
		return model.AIProvider{}, model.ErrNotFound
	}
	return p, nil
}

func (m *memProviders) Create(_ context.Context, p *model.AIProvider) error {
	p.ID = m.next
	m.next++
	m.byID[p.ID] = *p
	return nil
}

func (m *memProviders) Update(_ context.Context, p *model.AIProvider) error {
	m.byID[p.ID] = *p
	return nil
}

func (m *memProviders) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}

// --- mock ai.ModelRepository ---

type memModels struct {
	byID map[uint64]model.AIModel
	next uint64
}

func newMemModels() *memModels {
	return &memModels{byID: map[uint64]model.AIModel{}, next: 1}
}

func (m *memModels) FindByProviderID(_ context.Context, providerID uint64) ([]model.AIModel, error) {
	var result []model.AIModel
	for _, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			result = append(result, mdl)
		}
	}
	return result, nil
}

func (m *memModels) FindByID(_ context.Context, id uint64) (model.AIModel, error) {
	mdl, ok := m.byID[id]
	if !ok {
		return model.AIModel{}, model.ErrNotFound
	}
	return mdl, nil
}

func (m *memModels) Create(_ context.Context, mdl *model.AIModel) error {
	mdl.ID = m.next
	m.next++
	m.byID[mdl.ID] = *mdl
	return nil
}

func (m *memModels) Update(_ context.Context, mdl *model.AIModel) error {
	m.byID[mdl.ID] = *mdl
	return nil
}

func (m *memModels) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}

func (m *memModels) DeleteByProviderID(_ context.Context, providerID uint64) error {
	for id, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			delete(m.byID, id)
		}
	}
	return nil
}

// --- helpers ---

func validMasterKeyB64() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func newTestService(providers ai.ProviderRepository, models ai.ModelRepository, settings aigo.GameSettingRepository, client aigo.LLMClient) *aigo.Service {
	crypto, _ := security.NewCryptoManager(validMasterKeyB64())
	defaultAI := config.DefaultAIConfig{
		ProviderName: "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		ModelName:    "gpt-4o-mini",
		APIKey:       "sk-default",
	}
	return aigo.NewService(settings, providers, models, crypto, client, defaultAI, 30*time.Second)
}

func encryptForStorage(t *testing.T, plaintext string) string {
	t.Helper()
	crypto, _ := security.NewCryptoManager(validMasterKeyB64())
	enc, err := crypto.EncryptForStorage(plaintext)
	if err != nil {
		t.Fatalf("EncryptForStorage: %v", err)
	}
	return enc
}

func assertCode(t *testing.T, err error, expected int) {
	t.Helper()
	var appErr *response.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != expected {
		t.Fatalf("expected code %d, got %d: %s", expected, appErr.Code, appErr.Message)
	}
}

func simpleSnapshot() dto.AnalyzeRequest {
	return dto.AnalyzeRequest{
		Size:   3,
		Player: 1,
		Layout: dto.GoLayout{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
	}
}

// --- GetSetting tests ---

func TestGetSettingCreatesDefault(t *testing.T) {
	settings := newMemoryGameSettings()
	svc := newTestService(newMemProviders(), newMemModels(), settings, &mockLLMClient{})

	resp, err := svc.GetSetting(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if resp.ActiveModelID != 0 || !resp.IsDefault || !resp.AllowAIEndGame {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.ModelName != "gpt-4o-mini" {
		t.Fatalf("expected model name gpt-4o-mini, got %s", resp.ModelName)
	}
	if !resp.HasAPIKey {
		t.Fatal("expected HasAPIKey=true for default AI")
	}
}

func TestGetSettingAfterUpdate(t *testing.T) {
	settings := newMemoryGameSettings()
	svc := newTestService(newMemProviders(), newMemModels(), settings, &mockLLMClient{})

	allowEnd := false
	if _, err := svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{AllowAIEndGame: &allowEnd}); err != nil {
		t.Fatalf("UpdateSetting: %v", err)
	}

	resp, err := svc.GetSetting(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if resp.AllowAIEndGame {
		t.Fatal("expected AllowAIEndGame=false")
	}
}

// --- UpdateSetting tests ---

func TestUpdateSettingModelNotFound(t *testing.T) {
	settings := newMemoryGameSettings()
	svc := newTestService(newMemProviders(), newMemModels(), settings, &mockLLMClient{})

	modelID := uint64(999)
	_, err := svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{ActiveModelID: &modelID})
	if err == nil {
		t.Fatal("expected error for non-existent model")
	}
	assertCode(t, err, response.CodeNotFound)
}

func TestUpdateSettingModelDisabled(t *testing.T) {
	settings := newMemoryGameSettings()
	providers := newMemProviders()
	models := newMemModels()

	provider := model.AIProvider{UserID: 1, ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", APIKeyEncrypted: "enc", Status: model.StatusActive}
	providers.Create(context.Background(), &provider)
	mdl := model.AIModel{ProviderID: provider.ID, ModelName: "gpt-4o", Status: model.StatusDisabled}
	models.Create(context.Background(), &mdl)

	svc := newTestService(providers, models, settings, &mockLLMClient{})
	modelID := mdl.ID
	_, err := svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{ActiveModelID: &modelID})
	if err == nil {
		t.Fatal("expected error for disabled model")
	}
	assertCode(t, err, response.CodeModelInactive)
}

func TestUpdateSettingModelNotOwned(t *testing.T) {
	settings := newMemoryGameSettings()
	providers := newMemProviders()
	models := newMemModels()

	provider := model.AIProvider{UserID: 2, ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", APIKeyEncrypted: "enc", Status: model.StatusActive}
	providers.Create(context.Background(), &provider)
	mdl := model.AIModel{ProviderID: provider.ID, ModelName: "gpt-4o", Status: model.StatusActive}
	models.Create(context.Background(), &mdl)

	svc := newTestService(providers, models, settings, &mockLLMClient{})
	modelID := mdl.ID
	_, err := svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{ActiveModelID: &modelID})
	if err == nil {
		t.Fatal("expected error for model not owned by user")
	}
	assertCode(t, err, response.CodeNotFound)
}

// --- Analyze tests ---

func TestAnalyzeDefaultModelMove(t *testing.T) {
	settings := newMemoryGameSettings()
	client := &mockLLMClient{content: `{"action":"move","vertex":[5,5]}`}
	svc := newTestService(newMemProviders(), newMemModels(), settings, client)

	resp, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if resp.Action != "move" {
		t.Fatalf("expected action move, got %s", resp.Action)
	}
	if resp.Vertex == nil || resp.Vertex[0] != 5 || resp.Vertex[1] != 5 {
		t.Fatalf("expected vertex [5,5], got %v", resp.Vertex)
	}
}

func TestAnalyzeEndGameAllowed(t *testing.T) {
	settings := newMemoryGameSettings()
	client := &mockLLMClient{content: `{"action":"end_game"}`}
	svc := newTestService(newMemProviders(), newMemModels(), settings, client)

	resp, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if resp.Action != "end_game" {
		t.Fatalf("expected action end_game, got %s", resp.Action)
	}
	if resp.Vertex != nil {
		t.Fatal("expected nil vertex for end_game")
	}
}

func TestAnalyzeEndGameNotAllowedAfterUpdate(t *testing.T) {
	settings := newMemoryGameSettings()
	client := &mockLLMClient{content: `{"action":"end_game"}`}
	svc := newTestService(newMemProviders(), newMemModels(), settings, client)

	allowEnd := false
	svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{AllowAIEndGame: &allowEnd})

	_, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err == nil {
		t.Fatal("expected error when end_game not allowed")
	}
	assertCode(t, err, response.CodeAIResponseInvalid)
}

func TestAnalyzeLLMCallFailed(t *testing.T) {
	settings := newMemoryGameSettings()
	client := &mockLLMClient{err: errors.New("network error")}
	svc := newTestService(newMemProviders(), newMemModels(), settings, client)

	_, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err == nil {
		t.Fatal("expected error for LLM call failure")
	}
	assertCode(t, err, response.CodeAICallFailed)
}

func TestAnalyzeInvalidLLMResponse(t *testing.T) {
	settings := newMemoryGameSettings()
	client := &mockLLMClient{content: "not json"}
	svc := newTestService(newMemProviders(), newMemModels(), settings, client)

	_, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err == nil {
		t.Fatal("expected error for invalid LLM response")
	}
	assertCode(t, err, response.CodeAIResponseInvalid)
}

func TestAnalyzeInvalidSnapshot(t *testing.T) {
	settings := newMemoryGameSettings()
	svc := newTestService(newMemProviders(), newMemModels(), settings, &mockLLMClient{})

	snap := dto.AnalyzeRequest{
		Size:   3,
		Player: 1,
		Layout: dto.GoLayout{{0, 0}, {0, 0, 0}, {0, 0, 0}},
	}
	_, err := svc.Analyze(context.Background(), 1, snap)
	if err == nil {
		t.Fatal("expected error for invalid snapshot")
	}
	assertCode(t, err, response.CodeValidation)
}

func TestAnalyzeCustomModel(t *testing.T) {
	settings := newMemoryGameSettings()
	providers := newMemProviders()
	models := newMemModels()

	provider := model.AIProvider{UserID: 1, ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", APIKeyEncrypted: encryptForStorage(t, "sk-test"), Status: model.StatusActive}
	providers.Create(context.Background(), &provider)
	mdl := model.AIModel{ProviderID: provider.ID, ModelName: "gpt-4o", Status: model.StatusActive}
	models.Create(context.Background(), &mdl)

	client := &mockLLMClient{content: `{"action":"move","vertex":[3,3]}`}
	svc := newTestService(providers, models, settings, client)

	modelID := mdl.ID
	if _, err := svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{ActiveModelID: &modelID}); err != nil {
		t.Fatalf("UpdateSetting: %v", err)
	}

	resp, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if resp.Action != "move" {
		t.Fatalf("expected action move, got %s", resp.Action)
	}
}

func TestAnalyzeCustomModelDisabled(t *testing.T) {
	settings := newMemoryGameSettings()
	providers := newMemProviders()
	models := newMemModels()

	provider := model.AIProvider{UserID: 1, ProviderName: "OpenAI", BaseURL: "https://api.openai.com/v1", APIKeyEncrypted: "enc", Status: model.StatusActive}
	providers.Create(context.Background(), &provider)
	mdl := model.AIModel{ProviderID: provider.ID, ModelName: "gpt-4o", Status: model.StatusActive}
	models.Create(context.Background(), &mdl)

	svc := newTestService(providers, models, settings, &mockLLMClient{})
	modelID := mdl.ID
	svc.UpdateSetting(context.Background(), 1, aigo.UpdateSettingParams{ActiveModelID: &modelID})

	mdl.Status = model.StatusDisabled
	models.Update(context.Background(), &mdl)

	_, err := svc.Analyze(context.Background(), 1, simpleSnapshot())
	if err == nil {
		t.Fatal("expected error for disabled model")
	}
	assertCode(t, err, response.CodeModelInactive)
}
