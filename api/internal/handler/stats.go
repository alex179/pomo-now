package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// StatsHandler 处理统计相关的请求
type StatsHandler struct {
	sessionService *service.SessionService
	taskService    *service.TaskService
}

// NewStatsHandler 创建新的统计处理器，接收sessionService和taskService参数
func NewStatsHandler(sessionService *service.SessionService, taskService *service.TaskService) *StatsHandler {
	return &StatsHandler{
		sessionService: sessionService,
		taskService:    taskService,
	}
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

// HourlyStats 处理每小时统计请求
func (h *StatsHandler) HourlyStats(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取日期参数，默认为今天
	dateStr := r.URL.Query().Get("date")
	var targetDate time.Time
	if dateStr == "" {
		targetDate = time.Now()
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid date format")
			return
		}
	}

	hourlyStats, err := h.sessionService.GetHourlyStats(user.ID, targetDate)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, hourlyStats)
}

// DailyStats 处理每日统计请求
func (h *StatsHandler) DailyStats(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取日期范围参数
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr == "" || endStr == "" {
		// 默认显示本月数据
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	} else {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid start date format")
			return
		}
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid end date format")
			return
		}
	}

	dailyStats, err := h.sessionService.GetDailyStats(user.ID, start, end)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, dailyStats)
}

// CompletedTasks 处理已完成任务请求
func (h *StatsHandler) CompletedTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取日期参数，默认为今天
	dateStr := r.URL.Query().Get("date")
	var targetDate time.Time
	if dateStr == "" {
		targetDate = time.Now()
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid date format")
			return
		}
	}

	// 限制参数，默认显示前10个
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	tasks, err := h.taskService.GetCompletedTasksForDate(user.ID, targetDate, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, tasks)
}

// DayDetail 处理特定日期的详细统计
func (h *StatsHandler) DayDetail(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取日期参数
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		response.Error(w, http.StatusBadRequest, "date parameter required")
		return
	}

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid date format")
		return
	}

	dayDetail, err := h.sessionService.GetDayDetail(user.ID, targetDate)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, dayDetail)
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
	stats, err := h.sessionService.GetStats(user.ID, "week")
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, stats)
}

// getMonthStats 获取本月统计
func (h *StatsHandler) getMonthStats(w http.ResponseWriter, r *http.Request, user *model.User) {
	stats, err := h.sessionService.GetStats(user.ID, "month")
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(w, stats)
}
