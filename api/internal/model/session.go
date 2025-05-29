package model

import (
	"time"
)

// SessionType 表示会话类型
type SessionType string

const (
	SessionTypeFocus      SessionType = "focus"
	SessionTypeShortBreak SessionType = "short_break"
	SessionTypeLongBreak  SessionType = "long_break"
)

// Session 表示一个番茄钟会话
type Session struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	TaskID          *string     `json:"task_id,omitempty"`
	DurationMinutes int         `json:"duration_minutes"`
	Type            SessionType `json:"type"`
	StartedAt       time.Time   `json:"started_at"`
	EndedAt         *time.Time  `json:"ended_at,omitempty"`
}

// SessionCreate 表示创建会话的请求
type SessionCreate struct {
	TaskID          *string     `json:"task_id,omitempty"`
	DurationMinutes int         `json:"duration_minutes" validate:"required,min=1"`
	Type            SessionType `json:"type" validate:"required,oneof=focus short_break long_break"`
}

// SessionStats 表示会话统计
type SessionStats struct {
	TotalPomodoros    int `json:"total_pomodoros"`
	TotalMinutes      int `json:"total_minutes"`
	TotalFocusMinutes int `json:"total_focus_minutes"`
	TotalSessions     int `json:"total_sessions"`
	CompletedTasks    int `json:"completed_tasks"`
}

// Stats 表示API返回的统计数据
type Stats struct {
	TotalPomodoros int     `json:"total_pomodoros"`
	TotalMinutes   int     `json:"total_minutes"`
	CompletedTasks int     `json:"completed_tasks"`
	DailyAverage   float64 `json:"daily_average"`
	StreakDays     int     `json:"streak_days"`
	CompletionRate float64 `json:"completion_rate,omitempty"`
}

// DailyStats 表示每日统计
type DailyStats struct {
	Date              time.Time `json:"date"`
	FocusMinutes      int       `json:"focus_minutes"`
	CompletedSessions int       `json:"completed_sessions"`
	CompletedTasks    int       `json:"completed_tasks"`
}

// HourlyStats 表示每小时统计
type HourlyStats struct {
	Hour       int     `json:"hour"`
	HourStr    string  `json:"hour_str"`
	Sessions   int     `json:"sessions"`
	Minutes    int     `json:"minutes"`
	Percentage float64 `json:"percentage"`
}

// DayDetail 表示特定日期的详细统计
type DayDetail struct {
	Date           time.Time         `json:"date"`
	TotalPomodoros int               `json:"total_pomodoros"`
	TotalMinutes   int               `json:"total_minutes"`
	CompletedTasks int               `json:"completed_tasks"`
	HourlyStats    []*HourlyStats    `json:"hourly_stats"`
	TaskDetails    []*TaskCompletion `json:"task_details"`
}

// TaskCompletion 表示任务完成记录
type TaskCompletion struct {
	TaskID       string    `json:"task_id"`
	TaskTitle    string    `json:"task_title"`
	Pomodoros    int       `json:"pomodoros"`
	TotalMinutes int       `json:"total_minutes"`
	CompletedAt  time.Time `json:"completed_at"`
}
