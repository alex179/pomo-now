package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// SessionHandler 处理会话相关的请求
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler 创建新的会话处理器
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

// Sessions 处理会话列表和创建请求
func (h *SessionHandler) Sessions(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listSessions(w, r, user)
	case http.MethodPost:
		h.createSession(w, r, user)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// CurrentSession 处理当前会话的获取和更新请求
func (h *SessionHandler) CurrentSession(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCurrentSession(w, r, user)
	case http.MethodPut:
		h.updateCurrentSession(w, r, user)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listSessions 获取会话列表
func (h *SessionHandler) listSessions(w http.ResponseWriter, r *http.Request, user *model.User) {
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()
	sessions, err := h.sessionService.ListSessions(user.ID, startTime, endTime)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, sessions)
}

// createSession 创建新会话
func (h *SessionHandler) createSession(w http.ResponseWriter, r *http.Request, user *model.User) {
	var req model.SessionCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.sessionService.CreateSession(user.ID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(w, session)
}

// getCurrentSession 获取当前会话
func (h *SessionHandler) getCurrentSession(w http.ResponseWriter, r *http.Request, user *model.User) {
	session, err := h.sessionService.GetCurrentSession(user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, session)
}

// updateCurrentSession 更新当前会话
func (h *SessionHandler) updateCurrentSession(w http.ResponseWriter, r *http.Request, user *model.User) {
	// 这里假设前端会传递 sessionID
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		response.Error(w, http.StatusBadRequest, "session_id required")
		return
	}
	session, err := h.sessionService.EndSession(user.ID, req.SessionID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, session)
}
