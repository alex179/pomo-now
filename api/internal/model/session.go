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
	// StartedAt       time.Time   `json:"started_at"` // Removed as per schema
	EndedAt         time.Time   `json:"ended_at"` // Changed to non-nullable as per schema and application logic will set it
}

// RecordSessionRequest 表示记录会话的请求 (renamed from SessionCreate)
type RecordSessionRequest struct {
	TaskID          *string     `json:"task_id,omitempty"`
	DurationMinutes int         `json:"duration_minutes" validate:"required,min=1"`
	Type            SessionType `json:"type" validate:"required,oneof=focus short_break long_break"`
	EndedAt         *time.Time  `json:"ended_at,omitempty"` // Optional: if not provided, service sets to time.Now()
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

// UserStatistics represents aggregated statistics for a user over a period.
type UserStatistics struct {
	TotalFocusSessions  int       `json:"total_focus_sessions"`
	TotalFocusMinutes   int       `json:"total_focus_minutes"`
	TotalTasksCompleted int       `json:"total_tasks_completed"` // Option A: Count of tasks completed within the period
	Period              string    `json:"period"`                // "today", "week", "month"
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
}
