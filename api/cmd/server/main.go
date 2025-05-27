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

	// 创建处理器实例
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	statsHandler := handler.NewStatsHandler(sessionService)

	// 创建新的ServeMux
	mux := http.NewServeMux()

	// API路由
	apiMux := http.NewServeMux()

	// 认证路由
	apiMux.HandleFunc("POST /api/auth/register", authHandler.Register)
	apiMux.HandleFunc("POST /api/auth/login", authHandler.Login)
	apiMux.HandleFunc("POST /api/auth/apple", authHandler.AppleLogin)

	// 任务路由
	apiMux.HandleFunc("/api/tasks", taskHandler.Tasks)
	apiMux.HandleFunc("/api/tasks/{id}", taskHandler.Task)

	// 会话路由
	apiMux.HandleFunc("/api/sessions", sessionHandler.Sessions)
	apiMux.HandleFunc("/api/sessions/current", sessionHandler.CurrentSession)

	// 统计路由
	apiMux.HandleFunc("/api/stats", statsHandler.Stats)

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
