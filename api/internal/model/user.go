package model

import (
	"time"
)

// User 表示系统用户
type User struct {
	ID          string    `json:"id"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash *string   `json:"-"` // "-" to omit from JSON responses
	AppleUserID  *string   `json:"apple_user_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// User Preferences
	WorkDuration        *int `json:"work_duration,omitempty" db:"work_duration"`
	ShortBreakDuration  *int `json:"short_break_duration,omitempty" db:"short_break_duration"`
	LongBreakDuration   *int `json:"long_break_duration,omitempty" db:"long_break_duration"`
	LongBreakInterval   *int `json:"long_break_interval,omitempty" db:"long_break_interval"`
}

// UpdateUserPreferencesRequest defines the structure for updating user preferences.
// Fields are pointers to allow partial updates (only provided fields are changed).
type UpdateUserPreferencesRequest struct {
	WorkDuration        *int `json:"work_duration"`
	ShortBreakDuration  *int `json:"short_break_duration"`
	LongBreakDuration   *int `json:"long_break_duration"`
	LongBreakInterval   *int `json:"long_break_interval"`
}

// UserRegistration 表示用户注册请求
type UserRegistration struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserLogin 表示用户登录请求
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AppleLogin 表示Apple登录请求
type AppleLogin struct {
	IdentityToken string `json:"identity_token" validate:"required"`
	FullName      struct {
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"fullName,omitempty"`
}

// AuthResponse 表示认证响应
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ChangePasswordRequest defines the structure for changing a user's password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
