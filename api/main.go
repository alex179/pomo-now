package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex/pomo-now/internal/config"
	"github.com/alex/pomo-now/internal/handler"
	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
	"github.com/alex/pomo-now/internal/store"
)

func main() {
	// 初始化数据库
	db, err := store.NewStore()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化数据库表结构
	if err := db.InitSchema(); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	// 初始化服务
	authService := service.NewAuthService(db, "your-secret-key-here")

	// 创建OAuth配置
	oauthConfig := &config.OAuthConfig{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		AppleClientID:      os.Getenv("APPLE_CLIENT_ID"),
		ApplePrivateKey:    os.Getenv("APPLE_PRIVATE_KEY"),
		AppleKeyID:         os.Getenv("APPLE_KEY_ID"),
		AppleTeamID:        os.Getenv("APPLE_TEAM_ID"),
		RedirectURL:        os.Getenv("OAUTH_REDIRECT_URL"),
	}
	oauthService := service.NewOAuthService(oauthConfig, db)

	taskService := service.NewTaskService(db)
	sessionService := service.NewSessionService(db)
	settingsService := service.NewSettingsService(db)
	userService := service.NewUserService(db)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService, oauthService)
	taskHandler := handler.NewTaskHandler(taskService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	statsHandler := handler.NewStatsHandler(sessionService, taskService)
	settingsHandler := handler.NewSettingsHandler(settingsService, userService)

	// 创建多路复用器
	mux := http.NewServeMux()

	// CORS中间件
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// 允许的源列表
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://127.0.0.1:3000",
				"http://localhost:3001",
				"http://127.0.0.1:3001",
			}

			// 检查是否为允许的源
			allowOrigin := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					allowOrigin = true
					break
				}
			}

			// 如果没有匹配的源但有origin头，在开发模式下允许
			if !allowOrigin && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, map[string]string{"status": "ok"})
	})

	// 需要认证的路由
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// 认证相关路由（不需要认证的）
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/google", authHandler.GoogleLogin)
	mux.HandleFunc("POST /api/auth/apple", authHandler.AppleLogin)

	// 需要认证的用户相关路由
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(authHandler.GetCurrentUser)))
	mux.Handle("PUT /api/auth/me", authMiddleware(http.HandlerFunc(authHandler.UpdateUser)))
	mux.Handle("POST /api/auth/avatar", authMiddleware(http.HandlerFunc(authHandler.UploadAvatar)))

	// 任务路由
	mux.Handle("GET /api/tasks", authMiddleware(http.HandlerFunc(taskHandler.GetTasks)))
	mux.Handle("POST /api/tasks", authMiddleware(http.HandlerFunc(taskHandler.CreateTask)))
	mux.Handle("GET /api/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.GetTask)))
	mux.Handle("PUT /api/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.UpdateTask)))
	mux.Handle("DELETE /api/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.DeleteTask)))

	// 会话路由
	mux.Handle("GET /api/sessions", authMiddleware(http.HandlerFunc(sessionHandler.Sessions)))
	mux.Handle("GET /api/sessions/current", authMiddleware(http.HandlerFunc(sessionHandler.CurrentSession)))

	// 统计路由
	mux.Handle("GET /api/stats", authMiddleware(http.HandlerFunc(statsHandler.Stats)))
	mux.Handle("GET /api/stats/hourly", authMiddleware(http.HandlerFunc(statsHandler.HourlyStats)))
	mux.Handle("GET /api/stats/daily", authMiddleware(http.HandlerFunc(statsHandler.DailyStats)))
	mux.Handle("GET /api/stats/completed-tasks", authMiddleware(http.HandlerFunc(statsHandler.CompletedTasks)))
	mux.Handle("GET /api/stats/day-detail", authMiddleware(http.HandlerFunc(statsHandler.DayDetail)))

	// 设置路由 - 使用统一处理器处理不同HTTP方法
	mux.Handle("/api/settings/pomodoro", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetPomodoroSettings(w, r)
		case http.MethodPut:
			settingsHandler.UpdatePomodoroSettings(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	mux.Handle("/api/settings/notification", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetNotificationSettings(w, r)
		case http.MethodPut:
			settingsHandler.UpdateNotificationSettings(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	// 用户管理路由 - 使用统一处理器处理不同HTTP方法
	mux.Handle("/api/user/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetUserProfile(w, r)
		case http.MethodPut:
			settingsHandler.UpdateUserProfile(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	mux.Handle("POST /api/user/change-password", authMiddleware(http.HandlerFunc(settingsHandler.ChangePassword)))
	mux.Handle("POST /api/user/change-email", authMiddleware(http.HandlerFunc(settingsHandler.UpdateEmail)))

	// 应用CORS中间件
	handler := corsMiddleware(mux)

	// 创建服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
