package store

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/alex/pomo-now/internal/model"
	_ "github.com/go-sql-driver/mysql"
)

// Store 表示数据库存储
type Store struct {
	db   *sql.DB
	Task *TaskStore
}

// NewStore 创建新的数据库存储
func NewStore() (*Store, error) {
	// 从环境变量获取数据库配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnvOrDefault("DB_USER", "root"),
		getEnvOrDefault("DB_PASSWORD", "123456"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "3305"),
		getEnvOrDefault("DB_NAME", "pomo_now"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	taskStore := NewTaskStore(db)
	return &Store{db: db, Task: taskStore}, nil
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.db.Close()
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitSchema 初始化数据库表结构
func (s *Store) InitSchema() error {
	// 创建users表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE,
			username VARCHAR(50),
			password_hash VARCHAR(255),
			apple_user_id VARCHAR(255) UNIQUE,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT chk_auth_method CHECK (password_hash IS NOT NULL OR apple_user_id IS NOT NULL),
			INDEX idx_email (email),
			INDEX idx_apple_user_id (apple_user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// 添加缺失的字段到现有的users表
	s.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50)`)
	s.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`)

	// 创建user_profiles表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_profiles (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			username VARCHAR(50) NOT NULL,
			avatar VARCHAR(500),
			bio TEXT,
			timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
			language VARCHAR(10) DEFAULT 'zh-CN',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_id (user_id),
			INDEX idx_username (username)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_profiles table: %v", err)
	}

	// 创建pomodoro_settings表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS pomodoro_settings (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			work_minutes INT DEFAULT 25,
			short_break_minutes INT DEFAULT 5,
			long_break_minutes INT DEFAULT 15,
			long_break_interval INT DEFAULT 4,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_id (user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create pomodoro_settings table: %v", err)
	}

	// 创建notification_settings表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_settings (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			work_start_notification BOOLEAN DEFAULT TRUE,
			work_end_notification BOOLEAN DEFAULT TRUE,
			break_start_notification BOOLEAN DEFAULT TRUE,
			break_end_notification BOOLEAN DEFAULT TRUE,
			sound_enabled BOOLEAN DEFAULT TRUE,
			selected_sound VARCHAR(50) DEFAULT 'default',
			volume INT DEFAULT 70,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_id (user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notification_settings table: %v", err)
	}

	// 创建tasks表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			estimated_pomodoros INT DEFAULT 1,
			completed_pomodoros INT DEFAULT 0,
			status ENUM('pending', 'in_progress', 'completed') DEFAULT 'pending',
			color VARCHAR(20) DEFAULT 'red',
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP NULL,
			INDEX idx_user_id (user_id),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			INDEX idx_completed_at (completed_at)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %v", err)
	}

	// 创建pomodoro_sessions表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS pomodoro_sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			task_id VARCHAR(36),
			duration_minutes INT NOT NULL,
			type ENUM('focus', 'short_break', 'long_break') NOT NULL,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ended_at TIMESTAMP NULL,
			INDEX idx_user_id (user_id),
			INDEX idx_task_id (task_id),
			INDEX idx_started_at (started_at),
			INDEX idx_ended_at (ended_at)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create pomodoro_sessions table: %v", err)
	}

	return nil
}

// ListTasks 获取用户的任务列表
func (s *Store) ListTasks(userID string) ([]*model.Task, error) {
	return s.Task.GetUserTasks(userID)
}

// CreateTask 创建新任务
func (s *Store) CreateTask(userID string, req *model.TaskCreate) (*model.Task, error) {
	return s.Task.CreateTask(userID, req.Title, req.EstimatedPomodoros, req.Color, req.Notes)
}

// GetTask 获取单个任务
func (s *Store) GetTask(id string) (*model.Task, error) {
	return s.Task.GetTaskByID(id)
}

// UpdateTask 更新任务
func (s *Store) UpdateTask(id string, req *model.TaskUpdate) (*model.Task, error) {
	return s.Task.UpdateTask(id, req)
}

// DeleteTask 删除任务
func (s *Store) DeleteTask(id string) error {
	return s.Task.DeleteTask(id)
}

// GetCompletedTasksForDate 获取指定日期完成的任务
func (s *Store) GetCompletedTasksForDate(userID string, date time.Time, limit int) ([]*model.TaskCompletion, error) {
	return s.Task.GetCompletedTasksForDate(userID, date, limit)
}
