package dto

// RegisterRequest 是普通用户注册请求。
type RegisterRequest struct {
	Username string `json:"username"`
	// Password 是前端使用 RSA 公钥加密后 base64 编码的密码密文。
	Password string `json:"password"`
}

// LoginRequest 是用户名密码登录请求。
type LoginRequest struct {
	Username string `json:"username"`
	// Password 是前端使用 RSA 公钥加密后 base64 编码的密码密文。
	Password string `json:"password"`
}

// ChangePasswordRequest 是已登录用户修改密码请求。
type ChangePasswordRequest struct {
	// OldPassword 是前端使用 RSA 公钥加密后 base64 编码的原密码密文。
	OldPassword string `json:"old_password"`
	// NewPassword 是前端使用 RSA 公钥加密后 base64 编码的新密码密文。
	NewPassword string `json:"new_password"`
}
