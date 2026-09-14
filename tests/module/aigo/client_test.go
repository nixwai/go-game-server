package aigo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	aigo "github.com/nixwai/go-game-server/app/module/aigo"
)

func TestHTTPLLMClientDisablesThinkingAndLimitsCompletion(t *testing.T) {
	var request struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Thinking  struct {
			Type string `json:"type"`
		} `json:"thinking"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"move\",\"vertex\":[1,1]}"}}]}`))
	}))
	defer server.Close()

	client := aigo.NewHTTPLLMClient()
	content, err := client.ChatCompletion(context.Background(), aigo.ChatRequest{
		BaseURL:         server.URL,
		Model:           "deepseek-v4-flash",
		Messages:        []aigo.ChatMessage{{Role: "user", Content: "move"}},
		MaxTokens:       64,
		DisableThinking: true,
		Timeout:         time.Second,
	})
	if err != nil {
		t.Fatalf("ChatCompletion: %v", err)
	}
	if content == "" {
		t.Fatal("expected completion content")
	}
	if request.Model != "deepseek-v4-flash" || request.MaxTokens != 64 || request.Thinking.Type != "disabled" {
		t.Fatalf("unexpected request options: %+v", request)
	}
}
