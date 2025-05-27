package handler

import (
	"net/http"
	"time"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// StatsHandler 处理统计相关的请求
type StatsHandler struct {
	sessionService *service.SessionService
}

// NewStatsHandler 创建新的统计处理器，接收sessionService参数
func NewStatsHandler(sessionService *service.SessionService) *StatsHandler {
	return &StatsHandler{sessionService: sessionService}
}

// Stats 处理统计请求
func (h *StatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取时间范围参数
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today" // 默认显示今日统计
	}

	switch period {
	case "today":
		h.getTodayStats(w, r, user)
	case "week":
		h.getWeekStats(w, r, user)
	case "month":
		h.getMonthStats(w, r, user)
	default:
		response.Error(w, http.StatusBadRequest, "invalid period")
	}
}

// getTodayStats 获取今日统计
func (h *StatsHandler) getTodayStats(w http.ResponseWriter, r *http.Request, user *model.User) {
	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	stats, err := h.sessionService.GetSessionStats(user.ID, start, end)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, stats)
}

// getWeekStats 获取本周统计
func (h *StatsHandler) getWeekStats(w http.ResponseWriter, r *http.Request, user *model.User) {
	end := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.AddDate(0, 0, -7)
	stats, err := h.sessionService.GetDailyStats(user.ID, start, end)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, stats)
}

// getMonthStats 获取本月统计
func (h *StatsHandler) getMonthStats(w http.ResponseWriter, r *http.Request, user *model.User) {
	end := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.AddDate(0, 0, -30)
	stats, err := h.sessionService.GetDailyStats(user.ID, start, end)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, stats)
}
