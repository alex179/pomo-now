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
