// Package service 实现认证相关业务规则。
package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/repository"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// AuthService 编排用户注册、登录和当前用户查询。
type AuthService struct {
	// users 是用户数据访问接口，隔离 Service 与具体数据库实现。
	users repository.UserRepository
	// hasher 负责密码哈希和校验。
	hasher security.PasswordHasher
	// tokens 负责生成登录后的 JWT。
	tokens *security.TokenManager
}

// LoginResult 是登录成功后返回的令牌和用户信息。
type LoginResult struct {
	// Token 是客户端后续请求使用的 Bearer Access Token。
	Token string
	// User 是服务层使用的持久化用户实体，Handler 必须先转换为 DTO 后才能作为接口响应。
	User model.User
}

// NewAuthService 创建认证服务，并注入其外部依赖。
func NewAuthService(users repository.UserRepository, hasher security.PasswordHasher, tokens *security.TokenManager) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens}
}

// Register 校验并创建普通用户，任何客户端都不能通过此接口创建管理员。
func (s *AuthService) Register(ctx context.Context, username, password string) (model.User, error) {
	if err := validateUsername(username); err != nil {
		return model.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return model.User{}, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return model.User{}, response.NewError(response.CodeInternal, "internal server error", err)
	}
	user := model.User{Username: strings.TrimSpace(username), PasswordHash: hash, Role: model.RoleUser, Status: model.StatusActive}
	if err := s.users.Create(ctx, &user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return model.User{}, response.NewError(response.CodeConflict, "username already exists", err)
		}
		return model.User{}, response.NewError(response.CodeInternal, "internal server error", err)
	}
	return user, nil
}

// Login 校验用户名密码和用户状态，成功后签发 JWT。
func (s *AuthService) Login(ctx context.Context, username, password string) (LoginResult, error) {
	if err := validateUsername(username); err != nil {
		return LoginResult{}, err
	}
	user, err := s.users.FindByUsername(ctx, strings.TrimSpace(username))
	// 统一处理用户不存在、密码错误和禁用状态，避免泄露账号是否存在。
	if err != nil || user.Status != model.StatusActive || !s.hasher.Compare(password, user.PasswordHash) {
		return LoginResult{}, response.NewError(response.CodeAuthFailed, "invalid username or password", nil)
	}
	token, err := s.tokens.Generate(user)
	if err != nil {
		return LoginResult{}, response.NewError(response.CodeInternal, "internal server error", err)
	}
	return LoginResult{Token: token, User: user}, nil
}

// CurrentUser 根据 JWT 中的用户 ID 查询最新用户状态，避免仅信任令牌快照。
func (s *AuthService) CurrentUser(ctx context.Context, userID uint64) (model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user.Status != model.StatusActive {
		return model.User{}, response.NewError(response.CodeNotFound, "user not found", nil)
	}
	return user, nil
}

// usernamePattern 限制用户名只使用 ASCII 字母、数字和下划线。
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// validateUsername 校验用户名长度、字符集和可用格式。
func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 64 || !usernamePattern.MatchString(username) {
		return response.NewError(response.CodeValidation, "invalid username", nil)
	}
	return nil
}

// validatePassword 校验密码长度，并要求同时包含字母和数字。
func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return response.NewError(response.CodeValidation, "invalid password", nil)
	}
	var letter, digit bool
	for _, r := range password {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !letter || !digit {
		return response.NewError(response.CodeValidation, "invalid password", nil)
	}
	return nil
}
