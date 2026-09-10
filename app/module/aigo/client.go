package aigo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChatMessage 是发送给 LLM 的对话消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 封装调用 LLM 所需的全部参数。
type ChatRequest struct {
	BaseURL  string
	APIKey   string
	Model    string
	Messages []ChatMessage
	Timeout  time.Duration
}

// LLMClient 定义与 LLM 交互的抽象接口。
type LLMClient interface {
	ChatCompletion(ctx context.Context, req ChatRequest) (string, error)
}

// HTTPLLMClient 是基于 net/http 的 OpenAI 兼容 Chat Completions 客户端。
type HTTPLLMClient struct {
	client *http.Client
}

// NewHTTPLLMClient 创建 HTTP LLM 客户端。
func NewHTTPLLMClient() *HTTPLLMClient {
	return &HTTPLLMClient{client: &http.Client{}}
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ChatCompletion 向 OpenAI 兼容端点发送请求并返回文本内容。
func (c *HTTPLLMClient) ChatCompletion(ctx context.Context, req ChatRequest) (string, error) {
	body := chatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	url := req.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call LLM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status %d", resp.StatusCode)
	}

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode LLM response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("LLM API returned no choices")
	}
	return result.Choices[0].Message.Content, nil
}
