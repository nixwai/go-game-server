// Package authdto 定义认证接口的请求和响应数据传输对象。
package auth

// RegisterRequest 是普通用户注册请求。
type RegisterRequest struct {
	// Username 是用户登录名。
	Username string `json:"username"`
	// Password 是用户明文密码，仅在请求处理期间使用。
	Password string `json:"password"`
}

// LoginRequest 是用户名密码登录请求。
type LoginRequest struct {
	// Username 是用户登录名。
	Username string `json:"username"`
	// Password 是用户明文密码，仅在请求处理期间使用。
	Password string `json:"password"`
}
