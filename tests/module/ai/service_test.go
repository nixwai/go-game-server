package ai_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/model"
	ai "github.com/nixwai/go-game-server/app/module/ai"

	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// validMasterKeyB64 返回一个合法的 base64 编码 32 字节主密钥。
func validMasterKeyB64() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}

// encryptWithPublicKey 使用 RSA-OAEP + SHA-256 加密明文，模拟前端行为。
func encryptWithPublicKey(t *testing.T, pubKeyPEM string, plaintext string) string {
	t.Helper()
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		t.Fatal("failed to decode PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse public key: %v", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		t.Fatal("not RSA public key")
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte(plaintext), nil)
	if err != nil {
		t.Fatalf("RSA encrypt: %v", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// --- 内存仓储 mock ---

type memoryAIProviders struct {
	byID   map[uint64]model.AIProvider
	next   uint64
	err    error
	models *memoryAIModels
}

func newMemoryAIProviders() *memoryAIProviders {
	return &memoryAIProviders{byID: map[uint64]model.AIProvider{}, next: 1}
}
func (m *memoryAIProviders) FindByUserID(_ context.Context, userID uint64) ([]model.AIProvider, error) {
	var result []model.AIProvider
	for _, p := range m.byID {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *memoryAIProviders) FindByID(_ context.Context, id uint64) (model.AIProvider, error) {
	if m.err != nil {
		return model.AIProvider{}, m.err
	}
	p, ok := m.byID[id]
	if !ok {
		return model.AIProvider{}, model.ErrNotFound
	}
	return p, nil
}
func (m *memoryAIProviders) Create(_ context.Context, p *model.AIProvider) error {
	p.ID = m.next
	m.next++
	m.byID[p.ID] = *p
	return nil
}
func (m *memoryAIProviders) Update(_ context.Context, p *model.AIProvider) error {
	m.byID[p.ID] = *p
	return nil
}
func (m *memoryAIProviders) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}
func (m *memoryAIProviders) DeleteWithModels(_ context.Context, providerID uint64) error {
	delete(m.byID, providerID)
	if m.models != nil {
		m.models.DeleteByProviderID(context.Background(), providerID)
	}
	return nil
}

type memoryAIModels struct {
	byID map[uint64]model.AIModel
	next uint64
}

func newMemoryAIModels() *memoryAIModels {
	return &memoryAIModels{byID: map[uint64]model.AIModel{}, next: 1}
}
func (m *memoryAIModels) FindByProviderID(_ context.Context, providerID uint64) ([]model.AIModel, error) {
	var result []model.AIModel
	for _, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			result = append(result, mdl)
		}
	}
	return result, nil
}
func (m *memoryAIModels) FindByID(_ context.Context, id uint64) (model.AIModel, error) {
	mdl, ok := m.byID[id]
	if !ok {
		return model.AIModel{}, model.ErrNotFound
	}
	return mdl, nil
}
func (m *memoryAIModels) Create(_ context.Context, mdl *model.AIModel) error {
	mdl.ID = m.next
	m.next++
	m.byID[mdl.ID] = *mdl
	return nil
}
func (m *memoryAIModels) Update(_ context.Context, mdl *model.AIModel) error {
	m.byID[mdl.ID] = *mdl
	return nil
}
func (m *memoryAIModels) Delete(_ context.Context, id uint64) error {
	delete(m.byID, id)
	return nil
}
func (m *memoryAIModels) DeleteByProviderID(_ context.Context, providerID uint64) error {
	for id, mdl := range m.byID {
		if mdl.ProviderID == providerID {
			delete(m.byID, id)
		}
	}
	return nil
}

func newTestAIService(providers ai.ProviderRepository, models ai.ModelRepository) (*ai.Service, *security.CryptoManager) {
	if p, ok := providers.(*memoryAIProviders); ok {
		if mm, ok := models.(*memoryAIModels); ok {
			p.models = mm
		}
	}
	crypto, _ := security.NewCryptoManager(validMasterKeyB64())
	defaultAI := config.DefaultAIConfig{
		ProviderName: "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		ModelName:    "gpt-4o-mini",
		APIKey:       "sk-default",
	}
	return ai.NewService(providers, models, crypto, defaultAI), crypto
}

func TestCreateProvider(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-my-secret")

	provider, err := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if provider.ID == 0 {
		t.Fatal("provider ID should not be 0")
	}
	if provider.UserID != 1 {
		t.Fatalf("expected UserID 1, got %d", provider.UserID)
	}
	if provider.ProviderName != "OpenAI" {
		t.Fatalf("expected ProviderName OpenAI, got %s", provider.ProviderName)
	}
	if provider.APIKeyEncrypted == "" || provider.APIKeyEncrypted == "sk-my-secret" {
		t.Fatal("APIKeyEncrypted should be non-empty and not equal to plaintext")
	}
	if provider.Status != model.StatusActive {
		t.Fatalf("expected status active, got %s", provider.Status)
	}
}

func TestCreateProviderInvalidName(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	_, err := svc.CreateProvider(context.Background(), 1, "", "https://api.openai.com/v1", "enc")
	if err == nil {
		t.Fatal("empty name should fail")
	}
}

func TestCreateProviderDecryptFailure(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	_, err := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", "invalid-ciphertext")
	if err == nil {
		t.Fatal("invalid ciphertext should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeDecryptFailed {
		t.Fatalf("expected CodeDecryptFailed, got %v", err)
	}
}

func TestListProviders(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)

	list, err := svc.ListProviders(context.Background(), 1)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(list))
	}
}

func TestUpdateProvider(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-original")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)

	newName := "Anthropic"
	newURL := "https://api.anthropic.com"
	updated, err := svc.UpdateProvider(context.Background(), 1, provider.ID, ai.UpdateProviderParams{
		ProviderName: &newName,
		BaseURL:      &newURL,
	})
	if err != nil {
		t.Fatalf("update provider: %v", err)
	}
	if updated.ProviderName != "Anthropic" || updated.BaseURL != "https://api.anthropic.com" {
		t.Fatalf("unexpected provider: %+v", updated)
	}
	// API Key 应保留原值。
	if updated.APIKeyEncrypted != provider.APIKeyEncrypted {
		t.Fatal("API key should be unchanged")
	}
}

func TestUpdateProviderWithNewAPIKey(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey1 := encryptWithPublicKey(t, pubPEM, "sk-original")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey1)

	encKey2 := encryptWithPublicKey(t, pubPEM, "sk-new-key")
	updated, err := svc.UpdateProvider(context.Background(), 1, provider.ID, ai.UpdateProviderParams{
		EncryptedAPIKey: &encKey2,
	})
	if err != nil {
		t.Fatalf("update provider: %v", err)
	}
	if updated.APIKeyEncrypted == provider.APIKeyEncrypted {
		t.Fatal("API key should have changed")
	}
}

func TestUpdateDefaultProviderRejected(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	name := "NewName"
	_, err := svc.UpdateProvider(context.Background(), 1, model.DefaultAIProviderID, ai.UpdateProviderParams{
		ProviderName: &name,
	})
	if err == nil {
		t.Fatal("updating default provider should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeDefaultAIReadOnly {
		t.Fatalf("expected CodeDefaultAIReadOnly, got %v", err)
	}
}

func TestDeleteProvider(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)
	svc.CreateModel(context.Background(), 1, provider.ID, "gpt-4o")

	err := svc.DeleteProvider(context.Background(), 1, provider.ID)
	if err != nil {
		t.Fatalf("delete provider: %v", err)
	}
	// 关联模型应被级联删除。
	modelsList, _ := models.FindByProviderID(context.Background(), provider.ID)
	if len(modelsList) != 0 {
		t.Fatalf("expected 0 models after cascade delete, got %d", len(modelsList))
	}
}

func TestDeleteDefaultProviderRejected(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	err := svc.DeleteProvider(context.Background(), 1, model.DefaultAIProviderID)
	if err == nil {
		t.Fatal("deleting default provider should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeDefaultAIReadOnly {
		t.Fatalf("expected CodeDefaultAIReadOnly, got %v", err)
	}
}

func TestDeleteProviderWrongUser(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)

	err := svc.DeleteProvider(context.Background(), 2, provider.ID)
	if err == nil {
		t.Fatal("deleting another user's provider should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestCreateModel(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)

	mdl, err := svc.CreateModel(context.Background(), 1, provider.ID, "gpt-4o")
	if err != nil {
		t.Fatalf("create model: %v", err)
	}
	if mdl.ModelName != "gpt-4o" || mdl.ProviderID != provider.ID {
		t.Fatalf("unexpected model: %+v", mdl)
	}
}

func TestCreateModelWrongUser(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)

	_, err := svc.CreateModel(context.Background(), 2, provider.ID, "gpt-4o")
	if err == nil {
		t.Fatal("creating model on another user's provider should fail")
	}
}

func TestUpdateModel(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)
	mdl, _ := svc.CreateModel(context.Background(), 1, provider.ID, "gpt-4o")

	newName := "gpt-4o-mini"
	updated, err := svc.UpdateModel(context.Background(), 1, mdl.ID, ai.UpdateModelParams{ModelName: &newName})
	if err != nil {
		t.Fatalf("update model: %v", err)
	}
	if updated.ModelName != "gpt-4o-mini" {
		t.Fatalf("expected gpt-4o-mini, got %s", updated.ModelName)
	}
}

func TestUpdateDefaultModelRejected(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	name := "new-name"
	_, err := svc.UpdateModel(context.Background(), 1, model.DefaultAIModelID, ai.UpdateModelParams{ModelName: &name})
	if err == nil {
		t.Fatal("updating default model should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeDefaultAIReadOnly {
		t.Fatalf("expected CodeDefaultAIReadOnly, got %v", err)
	}
}

func TestDeleteModel(t *testing.T) {
	providers := newMemoryAIProviders()
	models := newMemoryAIModels()
	svc, crypto := newTestAIService(providers, models)

	pubPEM, _ := crypto.PublicKeyPEM()
	encKey := encryptWithPublicKey(t, pubPEM, "sk-secret")
	provider, _ := svc.CreateProvider(context.Background(), 1, "OpenAI", "https://api.openai.com/v1", encKey)
	mdl, _ := svc.CreateModel(context.Background(), 1, provider.ID, "gpt-4o")

	err := svc.DeleteModel(context.Background(), 1, mdl.ID)
	if err != nil {
		t.Fatalf("delete model: %v", err)
	}
}

func TestDeleteDefaultModelRejected(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	err := svc.DeleteModel(context.Background(), 1, model.DefaultAIModelID)
	if err == nil {
		t.Fatal("deleting default model should fail")
	}
	var appErr *response.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.CodeDefaultAIReadOnly {
		t.Fatalf("expected CodeDefaultAIReadOnly, got %v", err)
	}
}

func TestDefaultAI(t *testing.T) {
	svc, _ := newTestAIService(newMemoryAIProviders(), newMemoryAIModels())
	d := svc.DefaultAI()
	if d.ProviderName != "OpenAI" || d.ModelName != "gpt-4o-mini" {
		t.Fatalf("unexpected default AI: %+v", d)
	}
}
