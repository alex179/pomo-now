package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex/pomo-now/internal/handler"
	"github.com/alex/pomo-now/internal/middleware"
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
	authService := service.NewAuthService(dbStore, jwtSecret)
	taskService := service.NewTaskService(dbStore)
	sessionService := service.NewSessionService(dbStore)
	userService := service.NewUserService(dbStore) // Initialize UserService

	// 创建处理器实例
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	userHandler := handler.NewUserHandler(userService) // Initialize UserHandler
	// statsHandler := handler.NewStatsHandler(sessionService) // Commented out if NewStatsHandler and its methods are fully replaced by SessionHandler for these stats

	// 创建新的ServeMux
	mux := http.NewServeMux()

	// API路由
	apiMux := http.NewServeMux()

	// 创建新的ServeMux for main router
	mux := http.NewServeMux()

	// Public routes (no auth middleware)
	authRoutes := http.NewServeMux()
	authRoutes.HandleFunc("POST /api/auth/register", authHandler.Register)
	authRoutes.HandleFunc("POST /api/auth/login", authHandler.Login)
	authRoutes.HandleFunc("POST /api/auth/apple/login", authHandler.AppleLogin)

	// Apply general middlewares (logging, CORS) to public auth routes
	publicHandler := middleware.LoggingMiddleware(
		middleware.CORSMiddleware(authRoutes),
	)
	mux.Handle("/api/auth/", publicHandler) 

	// API routes that require authentication
	protectedApiMux := http.NewServeMux()
	// Add ChangePassword route to protectedApiMux as it requires authentication
	protectedApiMux.HandleFunc("POST /api/auth/change-password", authHandler.ChangePasswordHandler)
	// 任务路由
	protectedApiMux.HandleFunc("/api/tasks", taskHandler.Tasks)
	protectedApiMux.HandleFunc("/api/tasks/{id}", taskHandler.Task)
	// 会话路由
	protectedApiMux.HandleFunc("/api/sessions", sessionHandler.Sessions)
	// protectedApiMux.HandleFunc("/api/sessions/current", sessionHandler.CurrentSession) // Assuming CurrentSession routes are handled if still relevant, or removed if obsolete
	
	// 统计路由 - Changed to use sessionHandler.GetStatsHandler
	protectedApiMux.HandleFunc("GET /api/stats", sessionHandler.GetStatsHandler)

	// 用户偏好设置路由
	protectedApiMux.HandleFunc("GET /api/user/preferences", userHandler.GetPreferencesHandler)
	protectedApiMux.HandleFunc("PUT /api/user/preferences", userHandler.UpdatePreferencesHandler)


	// Apply all middlewares (logging, CORS, Auth) to protected API routes
	protectedApiHandler := middleware.LoggingMiddleware(
		middleware.CORSMiddleware(
			middleware.AuthMiddleware(authService)(protectedApiMux),
		),
	)
	// Use a more specific path for protected routes to avoid conflict if any non-/api/auth/ paths were meant to be public
	mux.Handle("/api/", protectedApiHandler)


	// 添加静态文件服务（仅在生产环境使用）
	// Note: This setup means static files are served under the main mux, without auth, logging, or CORS from the API middlewares.
	// If these middlewares are desired for static files too, this needs adjustment.
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
