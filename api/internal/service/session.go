package service

import (
	"errors"
	"time"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
)

// SessionService 处理会话相关的业务逻辑
type SessionService struct {
	store *store.Store
}

// NewSessionService 创建新的会话服务
func NewSessionService(store *store.Store) *SessionService {
	return &SessionService{
		store: store,
	}
}

// CreateSession 创建新会话
func (s *SessionService) CreateSession(userID string, req *model.SessionCreate) (*model.Session, error) {
	// 检查用户是否有活跃会话
	currentSession, err := s.store.GetCurrentSession(userID)
	if err != nil {
		return nil, err
	}
	if currentSession != nil {
		return nil, apperrors.ErrActiveSession
	}

	// 如果指定了任务，检查任务是否存在且属于当前用户
	if req.TaskID != nil {
		task, err := s.store.GetTask(*req.TaskID)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, apperrors.ErrTaskNotFound
		}
		if task.UserID != userID {
			return nil, apperrors.ErrNotAuthorized
		}
	}

	return s.store.CreateSession(userID, req)
}

// GetCurrentSession 获取当前会话
func (s *SessionService) GetCurrentSession(userID string) (*model.Session, error) {
	session, err := s.store.GetCurrentSession(userID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// ListSessions 获取用户的会话列表
func (s *SessionService) ListSessions(userID string, startTime, endTime time.Time) ([]*model.Session, error) {
	return s.store.ListSessions(userID, startTime, endTime)
}

// EndSession 结束会话
func (s *SessionService) EndSession(userID, sessionID string) (*model.Session, error) {
	// 获取会话
	session, err := s.store.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, apperrors.ErrSessionNotFound
	}
	if session.UserID != userID {
		return nil, apperrors.ErrNotAuthorized
	}
	if session.EndedAt != nil {
		return nil, errors.New("session already ended")
	}

	return s.store.EndSession(sessionID)
}

// GetSessionStats 获取会话统计
func (s *SessionService) GetSessionStats(userID string, startTime, endTime time.Time) (*model.SessionStats, error) {
	return s.store.GetSessionStats(userID, startTime, endTime)
}

// GetDailyStats 获取每日统计
func (s *SessionService) GetDailyStats(userID string, startTime, endTime time.Time) ([]*model.DailyStats, error) {
	return s.store.GetDailyStats(userID, startTime, endTime)
}

// GetStats 根据时间周期获取统计数据
func (s *SessionService) GetStats(userID, period string) (*model.Stats, error) {
	var startTime, endTime time.Time
	now := time.Now()

	switch period {
	case "today":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 0, 1)
	case "week":
		weekday := now.Weekday()
		if weekday == 0 {
			weekday = 7
		}
		startTime = now.AddDate(0, 0, -int(weekday-1))
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
		endTime = startTime.AddDate(0, 0, 7)
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 1, 0)
	default:
		return nil, errors.New("invalid period")
	}

	// 获取会话统计
	sessionStats, err := s.store.GetSessionStats(userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取任务完成数量
	completedTasks, err := s.store.GetCompletedTasksCount(userID, startTime, endTime)
	if err != nil {
		completedTasks = 0 // 如果出错，默认为0
	}

	// 计算任务完成率
	var completionRate float64
	// 获取该时间段内的所有任务数
	totalTasks, err := s.store.GetTasksCount(userID, startTime, endTime)
	if err == nil && totalTasks > 0 {
		completionRate = (float64(completedTasks) / float64(totalTasks)) * 100
	}

	// 计算平均值
	var dailyAverage float64
	if period == "week" {
		dailyAverage = float64(sessionStats.TotalPomodoros) / 7.0
	} else if period == "month" {
		daysInMonth := endTime.Sub(startTime).Hours() / 24
		dailyAverage = float64(sessionStats.TotalPomodoros) / daysInMonth
	} else {
		dailyAverage = float64(sessionStats.TotalPomodoros)
	}

	return &model.Stats{
		TotalPomodoros: sessionStats.TotalPomodoros,
		TotalMinutes:   sessionStats.TotalMinutes,
		CompletedTasks: completedTasks,
		DailyAverage:   dailyAverage,
		StreakDays:     0, // TODO: 实现连续天数计算
		CompletionRate: completionRate,
	}, nil
}

// GetHourlyStats 获取指定日期的每小时统计
func (s *SessionService) GetHourlyStats(userID string, date time.Time) ([]*model.HourlyStats, error) {
	return s.store.GetHourlyStats(userID, date)
}

// GetDayDetail 获取特定日期的详细统计
func (s *SessionService) GetDayDetail(userID string, date time.Time) (*model.DayDetail, error) {
	return s.store.GetDayDetail(userID, date)
}
