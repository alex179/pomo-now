package service

import (
	"context"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
	"github.com/google/uuid"
)

// SettingsService 设置服务
type SettingsService struct {
	store *store.Store
}

// NewSettingsService 创建设置服务
func NewSettingsService(store *store.Store) *SettingsService {
	return &SettingsService{store: store}
}

// GetPomodoroSettings 获取番茄钟设置
func (s *SettingsService) GetPomodoroSettings(ctx context.Context, userID string) (*model.PomodoroSettings, error) {
	settings, err := s.store.GetPomodoroSettings(ctx, userID)
	if err != nil {
		// 检查是否是StoreError类型且为not_found
		if storeErr, ok := err.(*store.StoreError); ok && storeErr.Code == "not_found" {
			return s.createDefaultPomodoroSettings(ctx, userID)
		}
		return nil, err
	}
	return settings, nil
}

// UpdatePomodoroSettings 更新番茄钟设置
func (s *SettingsService) UpdatePomodoroSettings(ctx context.Context, userID string, req *model.PomodoroSettingsRequest) (*model.PomodoroSettings, error) {
	settings := &model.PomodoroSettings{
		ID:                uuid.New().String(),
		UserID:            userID,
		WorkMinutes:       req.WorkMinutes,
		ShortBreakMinutes: req.ShortBreakMinutes,
		LongBreakMinutes:  req.LongBreakMinutes,
		LongBreakInterval: req.LongBreakInterval,
		UpdatedAt:         time.Now(),
	}

	// 检查是否已存在设置
	existing, err := s.store.GetPomodoroSettings(ctx, userID)
	if err == nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
	} else {
		settings.CreatedAt = time.Now()
	}

	return s.store.SavePomodoroSettings(ctx, settings)
}

// GetNotificationSettings 获取通知设置
func (s *SettingsService) GetNotificationSettings(ctx context.Context, userID string) (*model.NotificationSettings, error) {
	settings, err := s.store.GetNotificationSettings(ctx, userID)
	if err != nil {
		// 检查是否是StoreError类型且为not_found
		if storeErr, ok := err.(*store.StoreError); ok && storeErr.Code == "not_found" {
			return s.createDefaultNotificationSettings(ctx, userID)
		}
		return nil, err
	}
	return settings, nil
}

// UpdateNotificationSettings 更新通知设置
func (s *SettingsService) UpdateNotificationSettings(ctx context.Context, userID string, req *model.NotificationSettingsRequest) (*model.NotificationSettings, error) {
	settings := &model.NotificationSettings{
		ID:                     uuid.New().String(),
		UserID:                 userID,
		WorkStartNotification:  req.WorkStartNotification,
		WorkEndNotification:    req.WorkEndNotification,
		BreakStartNotification: req.BreakStartNotification,
		BreakEndNotification:   req.BreakEndNotification,
		SoundEnabled:           req.SoundEnabled,
		SelectedSound:          req.SelectedSound,
		Volume:                 req.Volume,
		UpdatedAt:              time.Now(),
	}

	// 检查是否已存在设置
	existing, err := s.store.GetNotificationSettings(ctx, userID)
	if err == nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
	} else {
		settings.CreatedAt = time.Now()
	}

	return s.store.SaveNotificationSettings(ctx, settings)
}

// createDefaultPomodoroSettings 创建默认番茄钟设置
func (s *SettingsService) createDefaultPomodoroSettings(ctx context.Context, userID string) (*model.PomodoroSettings, error) {
	settings := &model.PomodoroSettings{
		ID:                uuid.New().String(),
		UserID:            userID,
		WorkMinutes:       25,
		ShortBreakMinutes: 5,
		LongBreakMinutes:  15,
		LongBreakInterval: 4,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return s.store.SavePomodoroSettings(ctx, settings)
}

// createDefaultNotificationSettings 创建默认通知设置
func (s *SettingsService) createDefaultNotificationSettings(ctx context.Context, userID string) (*model.NotificationSettings, error) {
	settings := &model.NotificationSettings{
		ID:                     uuid.New().String(),
		UserID:                 userID,
		WorkStartNotification:  true,
		WorkEndNotification:    true,
		BreakStartNotification: true,
		BreakEndNotification:   true,
		SoundEnabled:           true,
		SelectedSound:          "default",
		Volume:                 70,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	return s.store.SaveNotificationSettings(ctx, settings)
}
