package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// UserHandler handles HTTP requests related to user preferences.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetPreferencesHandler handles requests to get user preferences.
// GET /api/user/preferences
func (h *UserHandler) GetPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	preferencesUser, err := h.userService.GetPreferences(r.Context(), user.ID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrUserNotFound:
				response.Error(w, http.StatusNotFound, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// We return a specific preferences response struct for clarity,
	// or we can return the relevant parts of the preferencesUser model.User.
	// For this task, returning the preference fields from the User model is fine.
	// To be explicit about the response structure:
	type PreferencesResponse struct {
		WorkDuration        *int `json:"work_duration,omitempty"`
		ShortBreakDuration  *int `json:"short_break_duration,omitempty"`
		LongBreakDuration   *int `json:"long_break_duration,omitempty"`
		LongBreakInterval   *int `json:"long_break_interval,omitempty"`
	}
	respData := PreferencesResponse{
		WorkDuration:       preferencesUser.WorkDuration,
		ShortBreakDuration: preferencesUser.ShortBreakDuration,
		LongBreakDuration:  preferencesUser.LongBreakDuration,
		LongBreakInterval:  preferencesUser.LongBreakInterval,
	}

	response.Success(w, respData)
}

// UpdatePreferencesHandler handles requests to update user preferences.
// PUT /api/user/preferences
func (h *UserHandler) UpdatePreferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.UpdateUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Ensure at least one field is provided for update
	if req.WorkDuration == nil && req.ShortBreakDuration == nil &&
		req.LongBreakDuration == nil && req.LongBreakInterval == nil {
		response.Error(w, http.StatusBadRequest, "at least one preference field must be provided for update")
		return
	}

	updatedUser, err := h.userService.UpdatePreferences(r.Context(), user.ID, req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrBadRequest:
				response.Error(w, http.StatusBadRequest, appErr.Message)
			case apperrors.ErrUserNotFound: // Should not happen if update was successful before final GetUserByID
				response.Error(w, http.StatusNotFound, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	
	// Similar to GetPreferencesHandler, respond with only the preference fields.
	type PreferencesResponse struct {
		WorkDuration        *int `json:"work_duration,omitempty"`
		ShortBreakDuration  *int `json:"short_break_duration,omitempty"`
		LongBreakDuration   *int `json:"long_break_duration,omitempty"`
		LongBreakInterval   *int `json:"long_break_interval,omitempty"`
	}
	respData := PreferencesResponse{
		WorkDuration:       updatedUser.WorkDuration,
		ShortBreakDuration: updatedUser.ShortBreakDuration,
		LongBreakDuration:  updatedUser.LongBreakDuration,
		LongBreakInterval:  updatedUser.LongBreakInterval,
	}

	response.Success(w, respData)
}
