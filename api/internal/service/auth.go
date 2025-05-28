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

	// 创建用户
	user, err := s.store.CreateEmailUser(req.Email, req.Password) // Changed to CreateEmailUser
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

	// 验证密码
	if !s.store.VerifyPassword(user, req.Password) {
		return nil, apperrors.ErrInvalidCredentials
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

// AppleLogin 处理Apple登录 (Sign in with Apple)
func (s *AuthService) AppleLogin(req *model.AppleLogin) (*model.AuthResponse, error) {
	// Mock Validation: Generate a mock apple_user_id from the identityToken.
	// In a real scenario, this involves validating the token with Apple and extracting the 'sub' claim.
	if req.IdentityToken == "" {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "identity token is required", nil)
	}
	mockAppleUserID := "mock_apple_user_id_from_" + req.IdentityToken
	// Optionally, could parse req.FullName if needed for new user creation, but not required by current task.

	// 查找或创建用户
	user, err := s.store.GetUserByAppleUserID(mockAppleUserID) // Corrected function name
	if err != nil {
		// Handle potential database errors, not just ErrUserNotFound
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "error fetching user", err)
	}

	if user == nil { // User does not exist, create new user
		newUser := &model.User{
			AppleUserID: &mockAppleUserID,
			// Email can be null, or extracted if the real Apple token validation provides it.
			// FullName could be used here if desired for other profile fields.
		}
		// The store.CreateUser function will generate ID and CreatedAt.
		// store.CreateAppleUser was a wrapper around store.CreateUser.
		// We can directly use store.CreateUser if we construct the user model here.
		// Or, ensure CreateAppleUser correctly sets the AppleUserID field and calls CreateUser.
		// The existing CreateAppleUser in store.go takes appleUserID string.
		user, err = s.store.CreateAppleUser(mockAppleUserID)
		if err != nil {
			return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "error creating apple user", err)
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

// ChangePassword handles the logic for changing a user's password.
func (s *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	// Retrieve the current user by userID
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return apperrors.NewAppError(apperrors.ErrInternalServer, "failed to retrieve user", err)
	}
	if user == nil {
		return apperrors.ErrUserNotFound
	}

	// Check if user has a password hash (e.g. user might have signed up with Apple ID only)
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return apperrors.NewAppError(apperrors.ErrBadRequest, "user does not have a password set up (e.g., registered via OAuth)", nil)
	}

	// Verify oldPassword against the stored user.PasswordHash
	// bcrypt.CompareHashAndPassword returns nil on success
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword))
	if err != nil {
		// If err is bcrypt.ErrMismatchedHashAndPassword, then it's an invalid old password
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return apperrors.ErrInvalidCredentials // Or a more specific "old password mismatch" error
		}
		// Other errors might be from bcrypt itself, treat as internal server error
		return apperrors.NewAppError(apperrors.ErrInternalServer, "error verifying old password", err)
	}

	// Validate newPassword (e.g., minimum length)
	// This should align with the model's validation tag, but service-level check is good too.
	if len(newPassword) < 8 { // Assuming min length of 8
		return apperrors.NewAppError(apperrors.ErrBadRequest, "new password is too short (minimum 8 characters)", nil)
	}
	// Optional: Check if newPassword is the same as oldPassword
	if newPassword == oldPassword {
		return apperrors.NewAppError(apperrors.ErrBadRequest, "new password must be different from the old password", nil)
	}


	// Hash the newPassword using bcrypt
	newHashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.NewAppError(apperrors.ErrInternalServer, "failed to hash new password", err)
	}
	newHashedPassword := string(newHashedPasswordBytes)

	// Call store.UpdatePassword(userID, newHashedPassword)
	err = s.store.UpdatePassword(userID, newHashedPassword)
	if err != nil {
		return apperrors.NewAppError(apperrors.ErrInternalServer, "failed to update password", err)
	}

	return nil // Success
}
