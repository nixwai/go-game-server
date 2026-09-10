// Package response 定义统一 HTTP 响应和业务错误。
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// CodeOK 表示业务处理成功。
	CodeOK = 0
	// CodeValidation 表示请求参数或数据校验失败。
	CodeValidation = 1001
	// CodeAuthFailed 表示用户名密码认证失败。
	CodeAuthFailed = 2001
	// CodeTokenInvalid 表示 JWT 缺失、无效或已过期。
	CodeTokenInvalid = 2002
	// CodeForbidden 表示当前用户没有访问权限。
	CodeForbidden = 3001
	// CodeConflict 表示资源状态冲突，例如用户名重复。
	CodeConflict = 4001
	// CodeNotFound 表示请求的资源不存在。
	CodeNotFound = 4004
	// CodeDefaultAIReadOnly 表示默认 AI 配置不可修改或删除。
	CodeDefaultAIReadOnly = 4005
	// CodeDecryptFailed 表示传输层 RSA 解密失败。
	CodeDecryptFailed = 4006
	// CodeModelInactive 表示激活的模型或产商已禁用。
	CodeModelInactive = 4007
	// CodeInternal 表示服务器内部异常。
	CodeInternal = 9000
	// CodeMasterKeyInvalid 表示加密主密钥未配置或无效。
	CodeMasterKeyInvalid = 9001
	// CodeAICallFailed 表示 LLM API 调用失败。
	CodeAICallFailed = 9002
	// CodeAIResponseInvalid 表示 LLM 返回内容无法解析为有效落子或结束申请。
	CodeAIResponseInvalid = 9003
)

// Body 是所有 API 响应使用的统一包装结构。
type Body struct {
	// Code 是业务码，成功固定为 0，非 0 表示具体错误类型。
	Code int `json:"code"`
	// Message 是面向客户端的简短提示，不包含内部错误细节。
	Message string `json:"message"`
	// Data 是成功响应的业务数据。
	Data any `json:"data,omitempty"`
}

// AppError 是可安全返回给客户端的业务错误。
type AppError struct {
	// Code 是该错误对应的业务码。
	Code int
	// Message 是对外展示的安全错误信息。
	Message string
	// Err 保存内部错误，仅用于服务端日志或错误链判断。
	Err error
}

// Error 实现 error 接口，同时保留内部错误链能力。
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewError 创建一个带业务码的应用错误。
func NewError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Write 写入统一格式的响应，HTTP 状态码始终为 200。
func Write(c *gin.Context, code int, message string, data any) {
	c.JSON(http.StatusOK, Body{Code: code, Message: message, Data: data})
}

// WriteError 将应用错误转换为安全的统一响应，HTTP 状态码始终为 200。
func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		Write(c, appErr.Code, appErr.Message, nil)
		return
	}
	Write(c, CodeInternal, "服务器内部错误", nil)
}
