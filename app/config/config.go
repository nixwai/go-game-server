// Package config 负责从环境变量加载并校验服务运行配置。
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 是服务运行所需的全部配置。
type Config struct {
	// AppEnv 表示当前运行环境，例如 development、test 或 production。
	AppEnv string
	// HTTPAddr 是 HTTP 服务监听地址。
	HTTPAddr string
	// MySQLDSN 是 GORM 使用的 MySQL 数据源名称。
	MySQLDSN string
	// JWTSecret 是签发和校验 HS256 JWT 的密钥。
	JWTSecret string
	// JWTIssuer 是 JWT 的签发方声明。
	JWTIssuer string
	// JWTExpiresIn 是 Access Token 的有效期。
	JWTExpiresIn time.Duration
	// Argon2Time 是 Argon2id 的时间成本参数。
	Argon2Time uint32
	// Argon2Memory 是 Argon2id 使用的内存成本，单位为 KiB。
	Argon2Memory uint32
	// Argon2Threads 是 Argon2id 使用的并行线程数。
	Argon2Threads uint8
	// Argon2KeyLen 是生成密码哈希的密钥长度，单位为字节。
	Argon2KeyLen uint32
	// Argon2SaltLen 是生成随机盐的长度，单位为字节。
	Argon2SaltLen uint32
	// MasterKey 是 base64 编码的 32 字节 AES-256 主密钥，用于 AI API Key 存储加密。
	MasterKey string
	// DefaultAI 是通过配置文件提供的只读默认 AI 产商和模型，不入数据库。
	DefaultAI DefaultAIConfig
	// GoLLMTimeout 是调用 LLM API 的超时时间。
	GoLLMTimeout time.Duration
}

// DefaultAIConfig 描述通过环境变量配置的默认 AI 产商和模型。
type DefaultAIConfig struct {
	// ProviderName 是默认产商名称。
	ProviderName string
	// BaseURL 是默认产商的 API 基础地址。
	BaseURL string
	// ModelName 是默认模型名称。
	ModelName string
	// APIKey 是默认产商的 API Key，直接存于内存，不入库。
	APIKey string
}

// Load 加载完整服务配置，并强制校验 JWT 密钥和加密主密钥。
func Load() (Config, error) { return load(true) }

// LoadForAdmin 加载管理员初始化命令所需配置。该命令不需要 JWT 密钥。
func LoadForAdmin() (Config, error) { return load(false) }

// load 根据 requireJWT 决定是否校验 JWT 配置，避免迁移和初始化命令依赖无关密钥。
func load(requireJWT bool) (Config, error) {
	if err := LoadDotEnv(); err != nil {
		return Config{}, err
	}
	expiresIn, err := durationEnv("JWT_EXPIRES_IN", 2*time.Hour)
	if err != nil {
		return Config{}, err
	}
	goLLMTimeout, err := durationEnv("GO_LLM_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	timeCost, err := uint32Env("ARGON2_TIME", 1)
	if err != nil {
		return Config{}, err
	}
	memory, err := uint32Env("ARGON2_MEMORY", 64*1024)
	if err != nil {
		return Config{}, err
	}
	threads, err := uint8Env("ARGON2_THREADS", 4)
	if err != nil {
		return Config{}, err
	}
	keyLen, err := uint32Env("ARGON2_KEY_LEN", 32)
	if err != nil {
		return Config{}, err
	}
	saltLen, err := uint32Env("ARGON2_SALT_LEN", 16)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:        env("APP_ENV", "development"),
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		MySQLDSN:      os.Getenv("MYSQL_DSN"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTIssuer:     env("JWT_ISSUER", "go-game-server"),
		JWTExpiresIn:  expiresIn,
		Argon2Time:    timeCost,
		Argon2Memory:  memory,
		Argon2Threads: threads,
		Argon2KeyLen:  keyLen,
		Argon2SaltLen: saltLen,
		MasterKey:     os.Getenv("MASTER_KEY"),
		GoLLMTimeout:  goLLMTimeout,
		DefaultAI: DefaultAIConfig{
			ProviderName: env("DEFAULT_AI_PROVIDER", "OpenAI"),
			BaseURL:      env("DEFAULT_AI_BASE_URL", "https://api.openai.com/v1"),
			ModelName:    env("DEFAULT_AI_MODEL", "gpt-4o-mini"),
			APIKey:       os.Getenv("DEFAULT_AI_API_KEY"),
		},
	}
	if cfg.MySQLDSN == "" {
		return Config{}, fmt.Errorf("MYSQL_DSN is required")
	}
	if requireJWT && len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	if cfg.Argon2Time == 0 || cfg.Argon2Memory < 19*1024 || cfg.Argon2Threads == 0 || cfg.Argon2KeyLen < 16 || cfg.Argon2SaltLen < 16 {
		return Config{}, fmt.Errorf("invalid Argon2 configuration")
	}
	if requireJWT {
		if cfg.MasterKey == "" {
			return Config{}, fmt.Errorf("MASTER_KEY is required")
		}
		masterBytes, err := base64.StdEncoding.DecodeString(cfg.MasterKey)
		if err != nil {
			return Config{}, fmt.Errorf("MASTER_KEY must be valid base64: %w", err)
		}
		if len(masterBytes) != 32 {
			return Config{}, fmt.Errorf("MASTER_KEY must decode to 32 bytes, got %d", len(masterBytes))
		}
	}
	return cfg, nil
}

// env 返回环境变量值；变量未设置时返回 fallback。
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// durationEnv 读取并解析 time.Duration 类型配置。
func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}

// uint32Env 读取无符号 32 位整数配置，并在格式错误时返回配置项名称。
func uint32Env(key string, fallback uint32) (uint32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return uint32(n), nil
}

// uint8Env 读取无符号 8 位整数配置，并在格式错误时返回配置项名称。
func uint8Env(key string, fallback uint8) (uint8, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return uint8(n), nil
}
