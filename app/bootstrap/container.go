// Package bootstrap 负责构建并运行 HTTP 服务应用。
package bootstrap

import (
	"fmt"

	"github.com/nixwai/go-game-server/app/config"
	"github.com/nixwai/go-game-server/app/security"
	"gorm.io/gorm"
)

// Deps 持有各业务模块运行所需的共享基础设施。
type Deps struct {
	// DB 是共享的 GORM 数据库连接。
	DB *gorm.DB
	// Config 是服务运行配置。
	Config config.Config
	// Tokens 负责生成和校验 JWT。
	Tokens *security.TokenManager
	// Crypto 负责 API Key 的加解密。
	Crypto *security.CryptoManager
}

// NewDeps 根据 DB 和配置构建共享运行时依赖。
func NewDeps(cfg config.Config, db *DB) (*Deps, error) {
	tokens := security.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiresIn)

	crypto, err := security.NewCryptoManager(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("create crypto manager: %w", err)
	}

	return &Deps{
		DB:     db.GORM,
		Config: cfg,
		Tokens: tokens,
		Crypto: crypto,
	}, nil
}
