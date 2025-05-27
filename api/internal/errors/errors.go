package errors

import "errors"

// 通用错误
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrNotAuthorized      = errors.New("not authorized")
)

// 任务相关错误
var (
	ErrTaskNotFound = errors.New("task not found")
)

// 会话相关错误
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrActiveSession   = errors.New("user has an active session")
)
