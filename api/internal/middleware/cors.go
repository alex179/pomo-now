package middleware

import "net/http"

// CORSMiddleware 处理跨域请求
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许的源
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// 设置允许的请求方法
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		// 设置允许的请求头
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		// 设置允许携带凭证
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// 设置预检请求的缓存时间
		w.Header().Set("Access-Control-Max-Age", "3600")
		// 设置暴露的响应头
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
