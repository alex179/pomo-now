package model

import (
	"database/sql"
	"fmt"
	"time"
)

// User 表示系统用户
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`      // 不在JSON响应中返回密码
	Avatar      string    `json:"avatar"` // 头像URL
	AppleUserID string    `json:"apple_user_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRegistration 表示用户注册请求
type UserRegistration struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserLogin 表示用户登录请求
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserUpdate 表示用户更新请求
type UserUpdate struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
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

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID        string       `json:"id"`
	Email     *string      `json:"email,omitempty"`
	Username  *string      `json:"username,omitempty"`
	IsActive  bool         `json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	Profile   *UserProfile `json:"profile,omitempty"`
}

func (u *User) Create(db *sql.DB) error {
	query := `
		INSERT INTO users (username, email, password, avatar, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	result, err := db.Exec(query, u.Username, u.Email, u.Password, u.Avatar)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = fmt.Sprintf("%d", id)
	return nil
}

func (u *User) Update(db *sql.DB) error {
	query := `
		UPDATE users 
		SET username = ?, email = ?, avatar = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := db.Exec(query, u.Username, u.Email, u.Avatar, u.ID)
	return err
}
