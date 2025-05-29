package model

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

type Task struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description,omitempty"`
	EstimatedPomodoros int        `json:"estimated_pomodoros"`
	CompletedPomodoros int        `json:"completed_pomodoros"`
	Status             TaskStatus `json:"status"`
	Color              string     `json:"color,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	Duration           int        `json:"duration"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

type TaskUpdate struct {
	Title              *string     `json:"title,omitempty"`
	EstimatedPomodoros *int        `json:"estimated_pomodoros,omitempty"`
	Status             *TaskStatus `json:"status,omitempty"`
}

// TaskCreate represents the payload for creating a new task
type TaskCreate struct {
	Title              string `json:"title"`
	EstimatedPomodoros int    `json:"estimated_pomodoros"`
	Color              string `json:"color,omitempty"`
	Notes              string `json:"notes,omitempty"`
}
