package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
	"github.com/google/uuid"
)

// AuthHandler 处理认证相关的请求
type AuthHandler struct {
	authService  *service.AuthService
	oauthService *service.OAuthService
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(authService *service.AuthService, oauthService *service.OAuthService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		oauthService: oauthService,
	}
}

// Register 处理用户注册
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req model.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	authResp, err := h.authService.Register(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(w, authResp)
}

// Login 处理用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	authResp, err := h.authService.Login(&req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "邮箱或密码错误")
		return
	}

	response.Success(w, authResp)
}

// AppleLogin 处理Apple登录
func (h *AuthHandler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req model.AppleLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 处理Apple登录
	user, err := h.oauthService.HandleAppleCallback("", req.IdentityToken)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Apple登录失败: "+err.Error())
		return
	}

	// 生成JWT token
	authResp, err := h.authService.Login(&model.UserLogin{
		Email:    user.Email,
		Password: "", // OAuth用户不需要密码验证
	})
	if err != nil {
		// 如果用户不存在，创建新用户
		if user.Email != "" {
			authResp, err = h.authService.Register(&model.UserRegistration{
				Email:    user.Email,
				Username: "Apple用户",
				Password: "", // OAuth用户不需要密码
			})
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "创建用户失败: "+err.Error())
				return
			}
		} else {
			response.Error(w, http.StatusBadRequest, "无法获取Apple用户信息")
			return
		}
	}

	response.Success(w, authResp)
}

// GoogleLogin 处理Google登录
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 处理Google登录
	user, err := h.oauthService.HandleGoogleCallback(req.Code)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Google登录失败: "+err.Error())
		return
	}

	// 生成JWT token
	authResp, err := h.authService.Login(&model.UserLogin{
		Email:    user.Email,
		Password: "", // OAuth用户不需要密码验证
	})
	if err != nil {
		// 如果用户不存在，创建新用户
		if user.Email != "" {
			authResp, err = h.authService.Register(&model.UserRegistration{
				Email:    user.Email,
				Username: "Google用户",
				Password: "", // OAuth用户不需要密码
			})
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "创建用户失败: "+err.Error())
				return
			}
		} else {
			response.Error(w, http.StatusBadRequest, "无法获取Google用户信息")
			return
		}
	}

	response.Success(w, authResp)
}

// GetGoogleAuthURL 获取Google登录URL
func (h *AuthHandler) GetGoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	authURL, state, err := h.oauthService.GetGoogleAuthURL()
	if err != nil {
		response.InternalError(w, err, "生成认证URL失败")
		return
	}

	response.Success(w, map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// GetAppleAuthURL 获取Apple登录URL
func (h *AuthHandler) GetAppleAuthURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	authURL, state, err := h.oauthService.GetAppleAuthURL()
	if err != nil {
		response.InternalError(w, err, "生成认证URL失败")
		return
	}

	response.Success(w, map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// Logout 处理用户退出登录
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 这里可以添加服务器端登录状态清理逻辑
	// 比如将token加入黑名单等
	// 目前简单返回成功，客户端负责清理本地存储

	response.Success(w, map[string]string{
		"message": "退出登录成功",
	})
}

// UploadAvatar 处理用户头像上传
func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	// 限制文件大小为5MB
	r.ParseMultipartForm(5 << 20)

	file, handler, err := r.FormFile("avatar")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "无法获取上传的文件")
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		response.Error(w, http.StatusBadRequest, "不支持的文件类型，请上传jpg、jpeg、png或gif格式的图片")
		return
	}

	// 创建上传目录
	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Error(w, http.StatusInternalServerError, "无法创建上传目录")
		return
	}

	// 生成唯一的文件名
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// 创建目标文件
	dst, err := os.Create(filepath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "无法创建文件")
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		response.Error(w, http.StatusInternalServerError, "无法保存文件")
		return
	}

	// 更新用户头像URL
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := h.authService.UpdateUserAvatar(user.ID, avatarURL); err != nil {
		response.Error(w, http.StatusInternalServerError, "无法更新用户头像")
		return
	}

	// 返回更新后的用户信息
	user.Avatar = avatarURL
	response.Success(w, user)
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	response.Success(w, user)
}

// UpdateUser 更新用户信息
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	updatedUser, err := h.authService.UpdateUser(user.ID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, updatedUser)
}
