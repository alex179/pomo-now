package errors

import (
	"encoding/json"
	"net/http"
)

// Error 表示API错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// New 创建一个新的API错误
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WriteJSON 将错误写入HTTP响应
func (e *Error) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(e)
}

// Common errors
var (
	ErrUnauthorized     = New(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden        = New(http.StatusForbidden, "Forbidden")
	ErrNotFound         = New(http.StatusNotFound, "Not Found")
	ErrBadRequest       = New(http.StatusBadRequest, "Bad Request")
	ErrInternalServer   = New(http.StatusInternalServerError, "Internal Server Error")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "Method Not Allowed")
)
