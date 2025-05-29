package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// SettingsHandler 处理设置相关的请求
type SettingsHandler struct {
	settingsService *service.SettingsService
	userService     *service.UserService
}

// NewSettingsHandler 创建新的设置处理器
func NewSettingsHandler(settingsService *service.SettingsService, userService *service.UserService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		userService:     userService,
	}
}

// GetPomodoroSettings 获取番茄钟设置
func (h *SettingsHandler) GetPomodoroSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	settings, err := h.settingsService.GetPomodoroSettings(r.Context(), user.ID)
	if err != nil {
		response.InternalError(w, err, "获取番茄钟设置失败")
		return
	}

	response.Success(w, settings)
}

// UpdatePomodoroSettings 更新番茄钟设置
func (h *SettingsHandler) UpdatePomodoroSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.PomodoroSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "请求格式错误")
		return
	}

	settings, err := h.settingsService.UpdatePomodoroSettings(r.Context(), user.ID, &req)
	if err != nil {
		response.InternalError(w, err, "更新番茄钟设置失败")
		return
	}

	response.Success(w, settings)
}

// GetNotificationSettings 获取通知设置
func (h *SettingsHandler) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	settings, err := h.settingsService.GetNotificationSettings(r.Context(), user.ID)
	if err != nil {
		response.InternalError(w, err, "获取通知设置失败")
		return
	}

	response.Success(w, settings)
}

// UpdateNotificationSettings 更新通知设置
func (h *SettingsHandler) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.NotificationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "请求格式错误")
		return
	}

	settings, err := h.settingsService.UpdateNotificationSettings(r.Context(), user.ID, &req)
	if err != nil {
		response.InternalError(w, err, "更新通知设置失败")
		return
	}

	response.Success(w, settings)
}

// ChangePassword 修改密码
func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "请求格式错误")
		return
	}

	err := h.userService.ChangePassword(r.Context(), user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "invalid current password" {
			response.ValidationError(w, "当前密码错误")
			return
		}
		response.InternalError(w, err, "密码修改失败")
		return
	}

	response.Success(w, map[string]string{
		"message": "密码修改成功",
	})
}

// GetUserProfile 获取用户资料
func (h *SettingsHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	userInfo, err := h.userService.GetUserInfo(r.Context(), user.ID)
	if err != nil {
		response.InternalError(w, err, "获取用户信息失败")
		return
	}

	response.Success(w, userInfo)
}

// UpdateUserProfile 更新用户资料
func (h *SettingsHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "请求格式错误")
		return
	}

	userInfo, err := h.userService.UpdateProfile(r.Context(), user.ID, &req)
	if err != nil {
		response.InternalError(w, err, "更新用户资料失败")
		return
	}

	response.Success(w, userInfo)
}

// UpdateEmail 修改邮箱
func (h *SettingsHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.UnauthorizedError(w)
		return
	}

	var req model.UpdateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "请求格式错误")
		return
	}

	err := h.userService.UpdateEmail(r.Context(), user.ID, req.NewEmail, req.Password)
	if err != nil {
		if err.Error() == "invalid password" {
			response.ValidationError(w, "密码错误")
			return
		}
		if err.Error() == "email already exists" {
			response.ConflictError(w, "该邮箱已被使用")
			return
		}
		response.InternalError(w, err, "邮箱修改失败")
		return
	}

	response.Success(w, map[string]string{
		"message": "邮箱修改成功",
	})
}
