package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// AuthHandler 处理认证相关的请求
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 处理用户注册
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(w, resp)
}

// Login 处理用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	response.Success(w, resp)
}

// ChangePasswordHandler handles requests to change a user's password.
// POST /api/auth/change-password
func (h *AuthHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // Or http.MethodPut
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Validate incoming data (basic presence checks, service layer handles more complex validation)
	if req.OldPassword == "" {
		response.Error(w, http.StatusBadRequest, "old_password is required")
		return
	}
	if req.NewPassword == "" {
		response.Error(w, http.StatusBadRequest, "new_password is required")
		return
	}
	// Min length for NewPassword is also enforced by model tag `validate:"min=8"`,
	// but an explicit check here can provide immediate feedback.
	// The service layer will also validate this.

	err := h.authService.ChangePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrBadRequest:
				response.Error(w, http.StatusBadRequest, appErr.Message)
			case apperrors.ErrInvalidCredentials:
				response.Error(w, http.StatusUnauthorized, appErr.Message) // Or 400 if preferred for "wrong old password"
			case apperrors.ErrUserNotFound: // Should ideally not happen if user is from context
				response.Error(w, http.StatusNotFound, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, map[string]string{"message": "Password changed successfully"}) // Or http.StatusNoContent
}

// AppleLogin 处理Apple登录
func (h *AuthHandler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.AppleLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.AppleLogin(&req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(w, resp)
}
