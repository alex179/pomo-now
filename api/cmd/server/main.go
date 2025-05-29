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
	// 创建数据库存储
	dbStore, err := store.NewStore()
	if err != nil {
		log.Fatalf("Failed to create database store: %v", err)
	}
	defer dbStore.Close()

	// 初始化数据库表结构
	if err := dbStore.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// 创建服务实例
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // 在生产环境中应该使用环境变量
	}

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

	authService := service.NewAuthService(dbStore, jwtSecret)
	oauthService := service.NewOAuthService(oauthConfig, dbStore)
	taskService := service.NewTaskService(dbStore)
	sessionService := service.NewSessionService(dbStore)
	settingsService := service.NewSettingsService(dbStore)
	userService := service.NewUserService(dbStore)

	// 创建处理器实例
	authHandler := handler.NewAuthHandler(authService, oauthService)
	taskHandler := handler.NewTaskHandler(taskService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	statsHandler := handler.NewStatsHandler(sessionService, taskService)
	settingsHandler := handler.NewSettingsHandler(settingsService, userService)

	// 创建新的ServeMux
	mux := http.NewServeMux()

	// API路由
	apiMux := http.NewServeMux()

	// 认证路由
	apiMux.HandleFunc("POST /api/auth/register", authHandler.Register)
	apiMux.HandleFunc("POST /api/auth/login", authHandler.Login)
	apiMux.HandleFunc("POST /api/auth/apple", authHandler.AppleLogin)
	apiMux.HandleFunc("POST /api/auth/google", authHandler.GoogleLogin)
	apiMux.HandleFunc("GET /api/auth/google/url", authHandler.GetGoogleAuthURL)
	apiMux.HandleFunc("GET /api/auth/apple/url", authHandler.GetAppleAuthURL)
	apiMux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	// 需要认证的路由
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// 任务路由
	apiMux.Handle("/api/tasks", authMiddleware(http.HandlerFunc(taskHandler.Tasks)))
	apiMux.Handle("/api/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.Task)))

	// 会话路由
	apiMux.Handle("/api/sessions", authMiddleware(http.HandlerFunc(sessionHandler.Sessions)))
	apiMux.Handle("/api/sessions/current", authMiddleware(http.HandlerFunc(sessionHandler.CurrentSession)))

	// 统计路由
	apiMux.Handle("/api/stats", authMiddleware(http.HandlerFunc(statsHandler.Stats)))
	apiMux.Handle("/api/stats/hourly", authMiddleware(http.HandlerFunc(statsHandler.HourlyStats)))
	apiMux.Handle("/api/stats/daily", authMiddleware(http.HandlerFunc(statsHandler.DailyStats)))
	apiMux.Handle("/api/stats/completed-tasks", authMiddleware(http.HandlerFunc(statsHandler.CompletedTasks)))
	apiMux.Handle("/api/stats/day-detail", authMiddleware(http.HandlerFunc(statsHandler.DayDetail)))

	// 设置路由
	apiMux.Handle("/api/settings/pomodoro", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetPomodoroSettings(w, r)
		case http.MethodPut:
			settingsHandler.UpdatePomodoroSettings(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	apiMux.Handle("/api/settings/notification", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetNotificationSettings(w, r)
		case http.MethodPut:
			settingsHandler.UpdateNotificationSettings(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	// 用户管理路由
	apiMux.Handle("/api/user/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settingsHandler.GetUserProfile(w, r)
		case http.MethodPut:
			settingsHandler.UpdateUserProfile(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})))

	apiMux.Handle("/api/user/change-password", authMiddleware(http.HandlerFunc(settingsHandler.ChangePassword)))
	apiMux.Handle("/api/user/change-email", authMiddleware(http.HandlerFunc(settingsHandler.UpdateEmail)))

	// 为API路由添加中间件
	apiHandler := middleware.LoggingMiddleware(
		middleware.CORSMiddleware(
			middleware.AuthMiddleware(authService)(apiMux),
		),
	)
	mux.Handle("/", apiHandler)

	// 添加静态文件服务（仅在生产环境使用）
	if os.Getenv("ENV") == "production" {
		fs := http.FileServer(http.Dir("../web/build"))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭服务器
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
