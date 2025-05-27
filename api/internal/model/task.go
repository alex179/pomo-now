package model

import (
	"time"
)

// TaskStatus 表示任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

// Task 表示一个任务
type Task struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Title              string     `json:"title"`
	EstimatedPomodoros int        `json:"estimated_pomodoros"`
	Status             TaskStatus `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

// TaskCreate 表示创建任务的请求
type TaskCreate struct {
	Title              string `json:"title" validate:"required"`
	EstimatedPomodoros int    `json:"estimated_pomodoros" validate:"required,min=1"`
}

// TaskUpdate 表示更新任务的请求
type TaskUpdate struct {
	Title              *string     `json:"title,omitempty"`
	EstimatedPomodoros *int        `json:"estimated_pomodoros,omitempty"`
	Status             *TaskStatus `json:"status,omitempty"`
}
