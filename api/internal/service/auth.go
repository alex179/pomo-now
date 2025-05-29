package service

import (
	"time"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService 处理认证相关的业务逻辑
type AuthService struct {
	store *store.Store
	// JWT密钥
	jwtSecret []byte
}

// NewAuthService 创建新的认证服务
func NewAuthService(store *store.Store, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register 处理用户注册
func (s *AuthService) Register(req *model.UserRegistration) (*model.AuthResponse, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.store.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperrors.ErrEmailExists
	}

	var user *model.User

	// 如果密码为空，说明是OAuth用户
	if req.Password == "" {
		user, err = s.store.CreateOAuthUser(req.Email, req.Username, "oauth")
	} else {
		user, err = s.store.CreateUser(req.Email, req.Password)
	}

	if err != nil {
		return nil, err
	}

	// 生成JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login 处理用户登录
func (s *AuthService) Login(req *model.UserLogin) (*model.AuthResponse, error) {
	// 获取用户
	user, err := s.store.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	// 对于OAuth用户（密码为空），跳过密码验证
	if req.Password != "" {
		// 验证密码
		if !s.store.VerifyPassword(user, req.Password) {
			return nil, apperrors.ErrInvalidCredentials
		}
	}

	// 生成JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// AppleLogin 处理Apple登录
func (s *AuthService) AppleLogin(req *model.AppleLogin) (*model.AuthResponse, error) {
	// TODO: 验证Apple identity token
	// 这里暂时使用identity token作为apple_user_id

	// 查找或创建用户
	user, err := s.store.GetUserByAppleID(req.IdentityToken)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = s.store.CreateAppleUser(req.IdentityToken)
		if err != nil {
			return nil, err
		}
	}

	// 生成JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// generateToken 生成JWT token
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken 验证JWT token
func (s *AuthService) ValidateToken(tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["sub"].(string)
		return s.store.GetUserByID(userID)
	}

	return nil, jwt.ErrSignatureInvalid
}

// UpdateUserAvatar 更新用户头像
func (s *AuthService) UpdateUserAvatar(userID string, avatarURL string) error {
	return s.store.UpdateUserAvatar(nil, userID, avatarURL)
}

// UpdateUser 更新用户信息
func (s *AuthService) UpdateUser(userID string, req *model.UserUpdate) (*model.User, error) {
	// 获取当前用户信息
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	// 如果更新邮箱，检查是否已存在
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.store.GetUserByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, apperrors.ErrEmailExists
		}
		user.Email = req.Email
	}

	// 更新其他字段
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// 保存更新
	err = s.store.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
