package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

// GetPomodoroSettings 获取番茄钟设置
func (s *Store) GetPomodoroSettings(ctx context.Context, userID string) (*model.PomodoroSettings, error) {
	query := `SELECT id, user_id, work_minutes, short_break_minutes, long_break_minutes, 
              long_break_interval, created_at, updated_at FROM pomodoro_settings WHERE user_id = ?`

	var settings model.PomodoroSettings
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.WorkMinutes,
		&settings.ShortBreakMinutes,
		&settings.LongBreakMinutes,
		&settings.LongBreakInterval,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &StoreError{Code: "not_found", Message: "settings not found"}
		}
		return nil, err
	}

	return &settings, nil
}

// SavePomodoroSettings 保存番茄钟设置
func (s *Store) SavePomodoroSettings(ctx context.Context, settings *model.PomodoroSettings) (*model.PomodoroSettings, error) {
	// 检查是否存在
	existing, err := s.GetPomodoroSettings(ctx, settings.UserID)
	if err != nil && err.(*StoreError).Code != "not_found" {
		return nil, err
	}

	if existing != nil {
		// 更新
		query := `UPDATE pomodoro_settings SET work_minutes = ?, short_break_minutes = ?, 
                  long_break_minutes = ?, long_break_interval = ?, updated_at = ? WHERE user_id = ?`
		_, err = s.db.ExecContext(ctx, query, settings.WorkMinutes, settings.ShortBreakMinutes,
			settings.LongBreakMinutes, settings.LongBreakInterval, time.Now(), settings.UserID)
		if err != nil {
			return nil, err
		}
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
	} else {
		// 创建
		settings.ID = uuid.New().String()
		settings.CreatedAt = time.Now()
		settings.UpdatedAt = time.Now()

		query := `INSERT INTO pomodoro_settings (id, user_id, work_minutes, short_break_minutes, 
                  long_break_minutes, long_break_interval, created_at, updated_at) 
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = s.db.ExecContext(ctx, query, settings.ID, settings.UserID, settings.WorkMinutes,
			settings.ShortBreakMinutes, settings.LongBreakMinutes, settings.LongBreakInterval,
			settings.CreatedAt, settings.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// GetNotificationSettings 获取通知设置
func (s *Store) GetNotificationSettings(ctx context.Context, userID string) (*model.NotificationSettings, error) {
	query := `SELECT id, user_id, work_start_notification, work_end_notification, 
              break_start_notification, break_end_notification, sound_enabled, 
              selected_sound, volume, created_at, updated_at 
              FROM notification_settings WHERE user_id = ?`

	var settings model.NotificationSettings
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.WorkStartNotification,
		&settings.WorkEndNotification,
		&settings.BreakStartNotification,
		&settings.BreakEndNotification,
		&settings.SoundEnabled,
		&settings.SelectedSound,
		&settings.Volume,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &StoreError{Code: "not_found", Message: "settings not found"}
		}
		return nil, err
	}

	return &settings, nil
}

// SaveNotificationSettings 保存通知设置
func (s *Store) SaveNotificationSettings(ctx context.Context, settings *model.NotificationSettings) (*model.NotificationSettings, error) {
	// 检查是否存在
	existing, err := s.GetNotificationSettings(ctx, settings.UserID)
	if err != nil && err.(*StoreError).Code != "not_found" {
		return nil, err
	}

	if existing != nil {
		// 更新
		query := `UPDATE notification_settings SET work_start_notification = ?, work_end_notification = ?, 
                  break_start_notification = ?, break_end_notification = ?, sound_enabled = ?, 
                  selected_sound = ?, volume = ?, updated_at = ? WHERE user_id = ?`
		_, err = s.db.ExecContext(ctx, query, settings.WorkStartNotification, settings.WorkEndNotification,
			settings.BreakStartNotification, settings.BreakEndNotification, settings.SoundEnabled,
			settings.SelectedSound, settings.Volume, time.Now(), settings.UserID)
		if err != nil {
			return nil, err
		}
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
	} else {
		// 创建
		settings.ID = uuid.New().String()
		settings.CreatedAt = time.Now()
		settings.UpdatedAt = time.Now()

		query := `INSERT INTO notification_settings (id, user_id, work_start_notification, work_end_notification, 
                  break_start_notification, break_end_notification, sound_enabled, selected_sound, volume, 
                  created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = s.db.ExecContext(ctx, query, settings.ID, settings.UserID, settings.WorkStartNotification,
			settings.WorkEndNotification, settings.BreakStartNotification, settings.BreakEndNotification,
			settings.SoundEnabled, settings.SelectedSound, settings.Volume, settings.CreatedAt, settings.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// StoreError 存储错误
type StoreError struct {
	Code    string
	Message string
}

func (e *StoreError) Error() string {
	return e.Message
}
