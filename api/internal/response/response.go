package response

import (
	"encoding/json"
	"net/http"
)

// JSON 发送JSON响应
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Error 发送错误响应
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{
		"error": message,
	})
}

// Success 发送成功响应
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created 发送创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent 发送无内容响应
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
