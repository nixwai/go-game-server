package ai

import (
	"net/url"
	"strings"

	"github.com/nixwai/go-game-server/app/response"
)

// validateProviderName 校验产商名称 trim 后长度在 1-128 之间。
func validateProviderName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 128 {
		return response.NewError(response.CodeValidation, "产商名称长度必须在 1-128 字符之间", nil)
	}
	return nil
}

// validateBaseURL 校验 Base URL 长度并要求 http/https 协议和有效主机。
func validateBaseURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if len(rawURL) < 1 || len(rawURL) > 512 {
		return response.NewError(response.CodeValidation, "Base URL 长度必须在 1-512 字符之间", nil)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return response.NewError(response.CodeValidation, "Base URL 格式无效", nil)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return response.NewError(response.CodeValidation, "Base URL 必须以 http 或 https 开头", nil)
	}
	if parsed.Host == "" {
		return response.NewError(response.CodeValidation, "Base URL 缺少主机地址", nil)
	}
	return nil
}

// validateModelName 校验模型名称 trim 后长度在 1-128 之间。
func validateModelName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 128 {
		return response.NewError(response.CodeValidation, "模型名称长度必须在 1-128 字符之间", nil)
	}
	return nil
}
