package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware 记录请求日志
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// 记录请求信息
		log.Printf(
			"%s %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter 是一个包装了http.ResponseWriter的结构体
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 重写WriteHeader方法以捕获状态码
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
