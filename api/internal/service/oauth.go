package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/alex/pomo-now/internal/config"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
)

// OAuthService OAuth服务
type OAuthService struct {
	config *config.OAuthConfig
	store  *store.Store
}

// GoogleUserInfo Google用户信息
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// AppleUserInfo Apple用户信息
type AppleUserInfo struct {
	Sub            string `json:"sub"`
	Email          string `json:"email"`
	EmailVerified  string `json:"email_verified"`
	AuthTime       int64  `json:"auth_time"`
	NonceSupported bool   `json:"nonce_supported"`
}

// NewOAuthService 创建新的OAuth服务
func NewOAuthService(config *config.OAuthConfig, store *store.Store) *OAuthService {
	return &OAuthService{
		config: config,
		store:  store,
	}
}

// GetGoogleAuthURL 获取Google登录URL
func (s *OAuthService) GetGoogleAuthURL() (string, string, error) {
	state, err := s.generateState()
	if err != nil {
		return "", "", err
	}

	params := url.Values{}
	params.Add("client_id", s.config.GoogleClientID)
	params.Add("redirect_uri", s.config.RedirectURL)
	params.Add("scope", "openid email profile")
	params.Add("response_type", "code")
	params.Add("state", state)

	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", params.Encode())
	return authURL, state, nil
}

// GetAppleAuthURL 获取Apple登录URL
func (s *OAuthService) GetAppleAuthURL() (string, string, error) {
	state, err := s.generateState()
	if err != nil {
		return "", "", err
	}

	params := url.Values{}
	params.Add("client_id", s.config.AppleClientID)
	params.Add("redirect_uri", s.config.RedirectURL)
	params.Add("scope", "name email")
	params.Add("response_type", "code")
	params.Add("response_mode", "form_post")
	params.Add("state", state)

	authURL := fmt.Sprintf("https://appleid.apple.com/auth/authorize?%s", params.Encode())
	return authURL, state, nil
}

// HandleGoogleCallback 处理Google回调
func (s *OAuthService) HandleGoogleCallback(code string) (*model.User, error) {
	// 获取访问令牌
	token, err := s.getGoogleAccessToken(code)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	userInfo, err := s.getGoogleUserInfo(token)
	if err != nil {
		return nil, err
	}

	// 查找或创建用户
	user, err := s.store.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// 创建新用户 - Google OAuth用户
		user, err = s.store.CreateOAuthUser(userInfo.Email, userInfo.Name, "google")
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// HandleAppleCallback 处理Apple回调
func (s *OAuthService) HandleAppleCallback(code string, idToken string) (*model.User, error) {
	// 验证ID Token (简化实现，生产环境需要完整验证)
	userInfo, err := s.parseAppleIDToken(idToken)
	if err != nil {
		return nil, err
	}

	// 查找或创建用户
	user, err := s.store.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// 创建新用户 - Apple OAuth用户
		user, err = s.store.CreateOAuthUser(userInfo.Email, "Apple用户", "apple")
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// generateState 生成随机状态字符串
func (s *OAuthService) generateState() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// getGoogleAccessToken 获取Google访问令牌
func (s *OAuthService) getGoogleAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.config.GoogleClientID)
	data.Set("client_secret", s.config.GoogleClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.config.RedirectURL)

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// getGoogleUserInfo 获取Google用户信息
func (s *OAuthService) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// parseAppleIDToken 解析Apple ID Token (简化实现)
func (s *OAuthService) parseAppleIDToken(idToken string) (*AppleUserInfo, error) {
	// 注意：这是简化实现，生产环境需要验证签名
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token format")
	}

	// 解码payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var userInfo AppleUserInfo
	if err := json.Unmarshal(payload, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
