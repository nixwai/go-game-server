package aigo_test

import (
	"context"
	"strings"
	"testing"

	aigo "github.com/nixwai/go-game-server/app/module/aigo"
)

type promptCaptureClient struct {
	request aigo.ChatRequest
}

func (c *promptCaptureClient) ChatCompletion(_ context.Context, req aigo.ChatRequest) (string, error) {
	c.request = req
	return `{"action":"move","vertex":[1,1]}`, nil
}

func TestAnalyzeUsesFastIntuitiveLLMSettings(t *testing.T) {
	client := &promptCaptureClient{}
	svc := newTestService(newMemProviders(), newMemModels(), newMemoryGameSettings(), client)

	if _, err := svc.Analyze(context.Background(), 1, simpleSnapshot()); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(client.request.Messages) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(client.request.Messages))
	}
	if client.request.MaxTokens != 64 {
		t.Fatalf("expected max tokens 64, got %d", client.request.MaxTokens)
	}
	if !client.request.DisableThinking {
		t.Fatal("expected thinking to be disabled")
	}
	systemPrompt := client.request.Messages[0].Content
	for _, phrase := range []string{"current board internally", "strongest and most effective legal move", "capture an opponent group", "save a group in danger", "Consider the last move and ko", "only JSON"} {
		if !strings.Contains(systemPrompt, phrase) {
			t.Fatalf("system prompt missing %q: %s", phrase, systemPrompt)
		}
	}
}
