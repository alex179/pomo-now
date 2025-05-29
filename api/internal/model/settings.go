package model

import (
	"time"
)

// PomodoroSettings 番茄钟设置
type PomodoroSettings struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	WorkMinutes       int       `json:"work_minutes"`        // 工作时长(分钟)
	ShortBreakMinutes int       `json:"short_break_minutes"` // 短休息时长(分钟)
	LongBreakMinutes  int       `json:"long_break_minutes"`  // 长休息时长(分钟)
	LongBreakInterval int       `json:"long_break_interval"` // 长休息间隔(几个番茄钟后)
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NotificationSettings 通知设置
type NotificationSettings struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	WorkStartNotification  bool      `json:"work_start_notification"`  // 工作开始通知
	WorkEndNotification    bool      `json:"work_end_notification"`    // 工作结束通知
	BreakStartNotification bool      `json:"break_start_notification"` // 休息开始通知
	BreakEndNotification   bool      `json:"break_end_notification"`   // 休息结束通知
	SoundEnabled           bool      `json:"sound_enabled"`            // 声音开启
	SelectedSound          string    `json:"selected_sound"`           // 选择的声音
	Volume                 int       `json:"volume"`                   // 音量(0-100)
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// UserProfile 用户资料
type UserProfile struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Avatar    *string   `json:"avatar,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	Timezone  string    `json:"timezone"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PomodoroSettingsRequest 番茄钟设置请求
type PomodoroSettingsRequest struct {
	WorkMinutes       int `json:"work_minutes" validate:"required,min=15,max=60"`
	ShortBreakMinutes int `json:"short_break_minutes" validate:"required,min=3,max=15"`
	LongBreakMinutes  int `json:"long_break_minutes" validate:"required,min=10,max=30"`
	LongBreakInterval int `json:"long_break_interval" validate:"required,min=2,max=10"`
}

// NotificationSettingsRequest 通知设置请求
type NotificationSettingsRequest struct {
	WorkStartNotification  bool   `json:"work_start_notification"`
	WorkEndNotification    bool   `json:"work_end_notification"`
	BreakStartNotification bool   `json:"break_start_notification"`
	BreakEndNotification   bool   `json:"break_end_notification"`
	SoundEnabled           bool   `json:"sound_enabled"`
	SelectedSound          string `json:"selected_sound" validate:"required"`
	Volume                 int    `json:"volume" validate:"min=0,max=100"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=2,max=50"`
	Bio      *string `json:"bio,omitempty" validate:"omitempty,max=200"`
	Timezone *string `json:"timezone,omitempty"`
	Language *string `json:"language,omitempty"`
}

// UpdateEmailRequest 修改邮箱请求
type UpdateEmailRequest struct {
	NewEmail string `json:"new_email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
