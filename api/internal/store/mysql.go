package store

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Store 表示数据库存储
type Store struct {
	db *sql.DB
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

	return &Store{db: db}, nil
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
			password_hash VARCHAR(255),
			apple_user_id VARCHAR(255) UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT chk_auth_method CHECK (password_hash IS NOT NULL OR apple_user_id IS NOT NULL)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// 创建tasks表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			estimated_pomodoros INT DEFAULT 1,
			status ENUM('pending', 'in_progress', 'completed') DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create pomodoro_sessions table: %v", err)
	}

	return nil
}
