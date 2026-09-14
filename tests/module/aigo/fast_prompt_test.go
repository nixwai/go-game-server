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
	for _, phrase := range []string{
		"current board internally",
		"strongest and most effective legal move",
		"capture an opponent group",
		"save a group in danger",
		"Consider the last move and ko",
		"Enumerate legal empty points",
		"A move is illegal if",
		"A point surrounded by enemy stones is not automatically playable",
		"re-check the exact cell",
		"only JSON",
	} {
		if !strings.Contains(systemPrompt, phrase) {
			t.Fatalf("system prompt missing %q: %s", phrase, systemPrompt)
		}
	}
	userPrompt := client.request.Messages[1].Content
	for _, phrase := range []string{
		"[x,y]",
		"layout[y][x]",
		"[1,2] means layout[2][1]",
		"Each row represents y",
		"left to right represent x",
		"x=0 1 2",
		"Empty points (candidate coordinates before suicide/ko checks): [0,0]",
		"Rule-filtered move points (non-suicide, excluding ko):",
		"Immediate capture points (if any):",
		"Forbidden empty points (suicide or ko; NEVER choose):",
		"FINAL MOVE CONSTRAINT",
		"FINAL SAFETY CHECK",
		"OUTPUT WHITELIST (machine-readable): {\"allowed_vertices\":[",
		"y=0 | . . .",
		"Before finalizing",
	} {
		if !strings.Contains(userPrompt, phrase) {
			t.Fatalf("user prompt missing %q: %s", phrase, userPrompt)
		}
	}
	if !strings.Contains(userPrompt, "y=2 | . X .") {
		t.Fatalf("expected layout[2][1] to be rendered at x=1, y=2: %s", userPrompt)
	}
}

func TestAnalyzePromptPrecomputesRuleFilteredPoints(t *testing.T) {
	client := &promptCaptureClient{}
	svc := newTestService(newMemProviders(), newMemModels(), newMemoryGameSettings(), client)
	snapshot := dto.AnalyzeRequest{
		Size: 5,
		Layout: dto.GoLayout{
			{0, 0, 0, 0, 0},
			{0, 0, -1, 0, 0},
			{0, -1, 0, -1, 0},
			{0, 0, -1, 0, 0},
			{0, 0, 0, 0, 0},
		},
		Player: 1,
	}

	if _, err := svc.Analyze(context.Background(), 1, snapshot); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	userPrompt := client.request.Messages[1].Content
	line := ""
	for _, candidate := range strings.Split(userPrompt, "\n") {
		if strings.HasPrefix(candidate, "Rule-filtered move points (non-suicide, excluding ko):") {
			line = candidate
			break
		}
	}
	if !strings.Contains(line, "[0,0]") || strings.Contains(line, "[2,2]") {
		t.Fatalf("expected suicide point to be removed from rule-filtered candidates: %s", userPrompt)
	}
	if !strings.Contains(userPrompt, "Forbidden empty points (suicide or ko; NEVER choose): [2,2]") {
		t.Fatalf("expected suicide point to be listed as forbidden: %s", userPrompt)
	}
}
