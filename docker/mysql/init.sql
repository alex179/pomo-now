-- Pomo-Now 数据库初始化脚本
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS pomo_now CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE pomo_now;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),
    apple_user_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建用户资料表
CREATE TABLE IF NOT EXISTS user_profiles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    display_name VARCHAR(255),
    bio TEXT,
    timezone VARCHAR(100) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_profiles_user_id (user_id)
);

-- 创建番茄钟设置表
CREATE TABLE IF NOT EXISTS pomodoro_settings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    work_minutes INT DEFAULT 25,
    short_break_minutes INT DEFAULT 5,
    long_break_minutes INT DEFAULT 15,
    long_break_interval INT DEFAULT 4,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pomodoro_settings_user_id (user_id)
);

-- 创建通知设置表
CREATE TABLE IF NOT EXISTS notification_settings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    work_start_notification BOOLEAN DEFAULT TRUE,
    work_end_notification BOOLEAN DEFAULT TRUE,
    break_start_notification BOOLEAN DEFAULT TRUE,
    break_end_notification BOOLEAN DEFAULT TRUE,
    sound_enabled BOOLEAN DEFAULT TRUE,
    selected_sound VARCHAR(100) DEFAULT 'default',
    volume INT DEFAULT 70,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_notification_settings_user_id (user_id)
);

-- 创建任务表
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_pomodoros INT DEFAULT 1,
    completed_pomodoros INT DEFAULT 0,
    status ENUM('pending', 'in_progress', 'completed') DEFAULT 'pending',
    color VARCHAR(7) DEFAULT '#FF5722',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    INDEX idx_tasks_user_id (user_id),
    INDEX idx_tasks_status (status),
    INDEX idx_tasks_created_at (created_at)
);

-- 创建番茄钟会话表
CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    task_id VARCHAR(36),
    session_type ENUM('work', 'short_break', 'long_break') DEFAULT 'work',
    planned_duration INT NOT NULL,
    actual_duration INT,
    completed BOOLEAN DEFAULT FALSE,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pomodoro_sessions_user_id (user_id),
    INDEX idx_pomodoro_sessions_task_id (task_id),
    INDEX idx_pomodoro_sessions_started_at (started_at)
);

-- 插入示例数据（可选）
-- 注意：在生产环境中可能不需要这些示例数据

-- 检查并显示创建的表
SHOW TABLES;

-- 输出初始化完成信息
SELECT 'Pomo-Now database initialized successfully!' AS status; 