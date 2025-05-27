package model

import (
	"time"
)

// User 表示系统用户
type User struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email,omitempty"`
	AppleUserID *string   `json:"apple_user_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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
