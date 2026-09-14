package aigo_test

import (
	"context"
	"strings"
	"testing"

	aigo "github.com/nixwai/go-game-server/app/module/aigo"
	"github.com/nixwai/go-game-server/app/module/aigo/dto"
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
	snapshot := simpleSnapshot()
	snapshot.Layout = dto.GoLayout{{0, 0, 0}, {0, 0, 0}, {0, 1, 0}}

	if _, err := svc.Analyze(context.Background(), 1, snapshot); err != nil {
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
	userPrompt := client.request.Messages[1].Content
	for _, phrase := range []string{"[x,y]", "layout[y][x]", "[1,2] means layout[2][1]", "Each row represents y", "left to right represent x"} {
		if !strings.Contains(userPrompt, phrase) {
			t.Fatalf("user prompt missing %q: %s", phrase, userPrompt)
		}
	}
	if !strings.Contains(userPrompt, "2: .X.") {
		t.Fatalf("expected layout[2][1] to be rendered at x=1, y=2: %s", userPrompt)
	}
}
