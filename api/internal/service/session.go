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
	// TaskService can be added here if task validation is needed, or passed as a param
}

// NewSessionService 创建新的会话服务
func NewSessionService(store *store.Store) *SessionService {
	return &SessionService{
		store: store,
	}
}

// RecordSession records a completed pomodoro session.
func (s *SessionService) RecordSession(ctx context.Context, userID string, req model.RecordSessionRequest) (*model.Session, error) {
	// Validate input
	if req.DurationMinutes <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "duration_minutes must be positive", nil)
	}

	isValidType := false
	for _, validType := range []model.SessionType{model.SessionTypeFocus, model.SessionTypeShortBreak, model.SessionTypeLongBreak} {
		if req.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "invalid session type", nil)
	}

	// Optional: TaskID validation (existence and ownership) can be added here if TaskService is available.
	// For now, per instructions, we'll just store it if provided.

	endedAt := time.Now().UTC()
	if req.EndedAt != nil {
		endedAt = req.EndedAt.UTC()
	}

	session := &model.Session{
		// ID will be set by the store
		UserID:          userID,
		TaskID:          req.TaskID,
		DurationMinutes: req.DurationMinutes,
		Type:            req.Type,
		EndedAt:         endedAt,
	}

	err := s.store.CreateSession(session)
	if err != nil {
		// TODO: Wrap error for more context? e.g., apperrors.NewAppError(apperrors.ErrInternalServer, "failed to record session", err)
		return nil, err
	}

	return session, nil
}


// GetCurrentSession 获取当前会话
// Note: This method was commented out in store/session.go due to schema changes.
// If it's needed, its logic and store counterpart would need to be re-evaluated.
/*
func (s *SessionService) GetCurrentSession(userID string) (*model.Session, error) {
	session, err := s.store.GetCurrentSession(userID)
	if err != nil {
		return nil, err
	}
	return session, nil
}
*/

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
	// if session.EndedAt != nil { // This check is against the old model where EndedAt was *time.Time
	// 	return nil, errors.New("session already ended")
	// }
	// The store.EndSession was also commented out. This method needs re-evaluation if "ending" a session becomes a distinct operation again.
	// For now, since sessions are recorded as completed, this method is likely obsolete.
	return nil, errors.New("EndSession functionality is currently obsolete due to schema changes")
}
*/
// GetSessionStats 获取会话统计
func (s *SessionService) GetSessionStats(userID string, startTime, endTime time.Time) (*model.SessionStats, error) {
	return s.store.GetSessionStats(userID, startTime, endTime)
}

// GetDailyStats 获取每日统计
func (s *SessionService) GetDailyStats(userID string, startTime, endTime time.Time) ([]*model.DailyStats, error) {
	return s.store.GetDailyStats(userID, startTime, endTime)
}

// GetStatistics calculates user statistics for a given period.
func (s *SessionService) GetStatistics(ctx context.Context, userID string, period string) (*model.UserStatistics, error) {
	now := time.Now().UTC()
	var startTime, endTime time.Time

	// Determine time range based on period
	switch period {
	case "today":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endTime = startTime.AddDate(0, 0, 1).Add(-time.Nanosecond)
	case "week":
		// Week starts on Monday
		weekday := now.Weekday()
		daysToSubtract := int(weekday) - int(time.Monday)
		if daysToSubtract < 0 {
			daysToSubtract += 7
		}
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysToSubtract)
		endTime = startTime.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		nextMonth := startTime.AddDate(0, 1, 0)
		endTime = nextMonth.Add(-time.Nanosecond)
	default:
		// Default to "today" if period is invalid or empty
		period = "today"
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endTime = startTime.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	totalFocusSessions, totalFocusMinutes, err := s.store.GetPomodoroStatsByUserID(userID, startTime, endTime)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to get pomodoro stats", err)
	}

	totalTasksCompleted, err := s.store.GetCompletedTasksCountByUserID(userID, startTime, endTime)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to get completed tasks count", err)
	}

	stats := &model.UserStatistics{
		TotalFocusSessions:  totalFocusSessions,
		TotalFocusMinutes:   totalFocusMinutes,
		TotalTasksCompleted: totalTasksCompleted,
		Period:              period,
		StartDate:           startTime,
		EndDate:             endTime,
	}

	return stats, nil
}
