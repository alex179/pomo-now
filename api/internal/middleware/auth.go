package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/service"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// 公开的认证端点列表
var publicAuthEndpoints = map[string]bool{
	"/api/auth/register": true,
	"/api/auth/login":    true,
	"/api/auth/apple":    true,
}

// AuthMiddleware 处理JWT认证
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否是公开的认证端点
			if publicAuthEndpoints[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
			token := parts[1]
			user, err := authService.ValidateToken(token)
			if err != nil || user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*model.User)
	return user, ok
}

// RequireAuth 确保请求已认证
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
