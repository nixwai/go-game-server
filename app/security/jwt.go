package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nixwai/go-game-server/app/model"
)

// Claims 是服务 JWT 中携带的自定义声明。
type Claims struct {
	// UserID 是用户主键，用于请求期间识别当前用户。
	UserID uint64 `json:"uid"`
	// Username 是签发令牌时的用户名快照。
	Username string `json:"username"`
	// Role 是签发令牌时的角色快照，用于快速进行权限判断。
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager 负责 JWT 的签发和校验。
type TokenManager struct {
	// secret 是 HS256 使用的签名密钥。
	secret []byte
	// issuer 是 JWT 的签发方，解析时会强制校验。
	issuer string
	// expiresIn 是 Access Token 的有效期。
	expiresIn time.Duration
}

// NewTokenManager 创建 JWT 管理器。
func NewTokenManager(secret, issuer string, expiresIn time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), issuer: issuer, expiresIn: expiresIn}
}

// Generate 为用户签发带有过期时间和唯一 ID 的 HS256 JWT。
func (m *TokenManager) Generate(user model.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 校验 JWT 的签名算法、签发方、有效期和令牌格式。
func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 限制签名算法为 HS256。
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
