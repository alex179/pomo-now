package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// APIResponse 标准API响应结构
type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON 发送JSON响应
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Success 发送成功响应
func Success(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Status: "success",
		Data:   data,
	}
	JSON(w, http.StatusOK, response)
}

// Created 发送创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Status: "success",
		Data:   data,
	}
	JSON(w, http.StatusCreated, response)
}

// Error 发送错误响应 - 安全的错误信息
func Error(w http.ResponseWriter, status int, message string) {
	ErrorWithCode(w, status, "GENERAL_ERROR", message)
}

// ErrorWithCode 发送带错误码的错误响应
func ErrorWithCode(w http.ResponseWriter, status int, code, message string) {
	response := APIResponse{
		Status: "error",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	JSON(w, status, response)
}

// InternalError 内部错误 - 记录详细日志但只返回通用错误给前端
func InternalError(w http.ResponseWriter, err error, userMessage string) {
	// 记录详细错误到日志
	log.Printf("Internal error: %v", err)

	// 返回通用错误给前端
	if userMessage == "" {
		userMessage = "服务暂时不可用，请稍后重试"
	}

	ErrorWithCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", userMessage)
}

// ValidationError 验证错误
func ValidationError(w http.ResponseWriter, message string) {
	ErrorWithCode(w, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

// NotFoundError 未找到错误
func NotFoundError(w http.ResponseWriter, resource string) {
	message := "请求的资源不存在"
	if resource != "" {
		message = resource + "不存在"
	}
	ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", message)
}

// UnauthorizedError 未授权错误
func UnauthorizedError(w http.ResponseWriter) {
	ErrorWithCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
}

// ForbiddenError 禁止访问错误
func ForbiddenError(w http.ResponseWriter) {
	ErrorWithCode(w, http.StatusForbidden, "FORBIDDEN", "权限不足")
}

// ConflictError 冲突错误
func ConflictError(w http.ResponseWriter, message string) {
	ErrorWithCode(w, http.StatusConflict, "CONFLICT", message)
}

// NoContent 发送无内容响应
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
