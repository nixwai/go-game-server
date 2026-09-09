package auth

// RegisterRequest 是普通用户注册请求。
type RegisterRequest struct {
	// Username 是用户登录名。
	Username string `json:"username"`
	// Password 是用户明文密码。
	Password string `json:"password"`
}

// LoginRequest 是用户名密码登录请求。
type LoginRequest struct {
	// Username 是用户登录名。
	Username string `json:"username"`
	// Password 是用户明文密码。
	Password string `json:"password"`
}
