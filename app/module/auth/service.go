package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/nixwai/go-game-server/app/model"
	"github.com/nixwai/go-game-server/app/response"
	"github.com/nixwai/go-game-server/app/security"
)

// Service 编排用户注册、登录和当前用户查询。
type Service struct {
	users  UserRepository
	hasher security.PasswordHasher
	tokens *security.TokenManager
}

// LoginResult 是登录成功后返回的令牌和用户信息。
type LoginResult struct {
	Token string
	User  model.User
}

// NewService 创建认证服务，并注入其外部依赖。
func NewService(users UserRepository, hasher security.PasswordHasher, tokens *security.TokenManager) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

// Register 校验并创建普通用户。
func (s *Service) Register(ctx context.Context, username, password string) (model.User, error) {
	if err := validateUsername(username); err != nil {
		return model.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return model.User{}, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return model.User{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	user := model.User{Username: strings.TrimSpace(username), PasswordHash: hash, Role: model.RoleUser, Status: model.StatusActive}
	if err := s.users.Create(ctx, &user); err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			return model.User{}, response.NewError(response.CodeConflict, "用户名已存在", err)
		}
		return model.User{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return user, nil
}

// Login 校验用户名密码和用户状态，成功后签发 JWT。
func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	if err := validateUsername(username); err != nil {
		return LoginResult{}, err
	}
	user, err := s.users.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil || user.Status != model.StatusActive || !s.hasher.Compare(password, user.PasswordHash) {
		return LoginResult{}, response.NewError(response.CodeAuthFailed, "用户名或密码错误", nil)
	}
	token, err := s.tokens.Generate(user)
	if err != nil {
		return LoginResult{}, response.NewError(response.CodeInternal, "服务器内部错误", err)
	}
	return LoginResult{Token: token, User: user}, nil
}

// CurrentUser 根据用户 ID 查询当前用户信息。
func (s *Service) CurrentUser(ctx context.Context, userID uint64) (model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user.Status != model.StatusActive {
		return model.User{}, response.NewError(response.CodeNotFound, "用户不存在", nil)
	}
	return user, nil
}

// usernamePattern 限制用户名只使用 ASCII 字母、数字和下划线。
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// validateUsername 校验用户名长度、字符集和可用格式。
func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 64 || !usernamePattern.MatchString(username) {
		return response.NewError(response.CodeValidation, "用户名格式无效", nil)
	}
	return nil
}

// validatePassword 校验密码长度，并要求同时包含字母和数字。
func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return response.NewError(response.CodeValidation, "密码格式无效", nil)
	}
	var letter, digit bool
	for _, r := range password {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !letter || !digit {
		return response.NewError(response.CodeValidation, "密码格式无效", nil)
	}
	return nil
}
