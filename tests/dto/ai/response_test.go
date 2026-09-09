package ai_test

import (
	"testing"
	"time"

	ai "github.com/nixwai/go-game-server/app/dto/ai"
	"github.com/nixwai/go-game-server/app/model"
)

func TestNewModelResponseNotDefault(t *testing.T) {
	now := time.Now()
	mdl := model.AIModel{ID: 5, ProviderID: 1, ModelName: "gpt-4o", Status: model.StatusActive, CreatedAt: now}
	resp := ai.NewModelResponse(mdl, false)
	if resp.ID != 5 || resp.ModelName != "gpt-4o" || resp.Status != "active" || resp.IsDefault {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestNewModelResponseDefault(t *testing.T) {
	now := time.Now()
	mdl := model.AIModel{ID: model.DefaultAIModelID, ModelName: "gpt-4o-mini", Status: model.StatusActive, CreatedAt: now}
	resp := ai.NewModelResponse(mdl, true)
	if !resp.IsDefault {
		t.Fatal("expected IsDefault=true")
	}
}

func TestNewProviderResponseDoesNotLeakAPIKey(t *testing.T) {
	now := time.Now()
	provider := model.AIProvider{
		ID:              3,
		UserID:          1,
		ProviderName:    "OpenAI",
		BaseURL:         "https://api.openai.com/v1",
		APIKeyEncrypted: "encrypted-secret-data",
		Status:          model.StatusActive,
		CreatedAt:       now,
	}
	resp := ai.NewProviderResponse(provider, nil, false, true)
	if resp.ID != 3 || resp.ProviderName != "OpenAI" || resp.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.HasAPIKey != true {
		t.Fatal("expected HasAPIKey=true")
	}
	if resp.IsDefault {
		t.Fatal("expected IsDefault=false for non-default provider")
	}
	// 确保响应中不包含 APIKeyEncrypted 字段——ProviderResponse 结构体中不存在该字段。
}

func TestNewProviderResponseDefault(t *testing.T) {
	now := time.Now()
	provider := model.AIProvider{
		ID:           model.DefaultAIProviderID,
		ProviderName: "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		Status:       model.StatusActive,
		CreatedAt:    now,
	}
	resp := ai.NewProviderResponse(provider, nil, true, true)
	if !resp.IsDefault {
		t.Fatal("expected IsDefault=true")
	}
	if resp.ID != 0 {
		t.Fatalf("expected ID=0 for default provider, got %d", resp.ID)
	}
}

func TestNewProviderResponseWithModels(t *testing.T) {
	now := time.Now()
	provider := model.AIProvider{ID: 1, ProviderName: "OpenAI", Status: "active", CreatedAt: now}
	models := []ai.ModelResponse{
		{ID: 1, ModelName: "gpt-4o", Status: "active"},
		{ID: 2, ModelName: "gpt-4o-mini", Status: "active"},
	}
	resp := ai.NewProviderResponse(provider, models, false, true)
	if len(resp.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(resp.Models))
	}
}
