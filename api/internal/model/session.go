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
	TotalFocusMinutes int `json:"total_focus_minutes"`
	TotalSessions     int `json:"total_sessions"`
	CompletedTasks    int `json:"completed_tasks"`
}

// DailyStats 表示每日统计
type DailyStats struct {
	Date              time.Time `json:"date"`
	FocusMinutes      int       `json:"focus_minutes"`
	CompletedSessions int       `json:"completed_sessions"`
	CompletedTasks    int       `json:"completed_tasks"`
}
