package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

// CreateSession 创建新会话
func (s *Store) CreateSession(userID string, req *model.SessionCreate) (*model.Session, error) {
	session := &model.Session{
		ID:              uuid.New().String(),
		UserID:          userID,
		TaskID:          req.TaskID,
		DurationMinutes: req.DurationMinutes,
		Type:            req.Type,
		StartedAt:       time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO pomodoro_sessions (id, user_id, task_id, duration_minutes, type, started_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.TaskID, session.DurationMinutes, session.Type, session.StartedAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession 获取单个会话
func (s *Store) GetSession(id string) (*model.Session, error) {
	var session model.Session
	var endedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, started_at, ended_at
		FROM pomodoro_sessions
		WHERE id = ?
	`, id).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.StartedAt, &endedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	return &session, nil
}

// GetCurrentSession 获取用户的当前会话
func (s *Store) GetCurrentSession(userID string) (*model.Session, error) {
	var session model.Session
	var endedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, started_at, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, userID).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.StartedAt, &endedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	return &session, nil
}

// ListSessions 获取用户的会话列表
func (s *Store) ListSessions(userID string, startTime, endTime time.Time) ([]*model.Session, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, task_id, duration_minutes, type, started_at, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? AND started_at BETWEEN ? AND ?
		ORDER BY started_at DESC
	`, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var session model.Session
		var endedAt sql.NullTime
		err := rows.Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.StartedAt, &endedAt)
		if err != nil {
			return nil, err
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, &session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// EndSession 结束会话
func (s *Store) EndSession(id string) (*model.Session, error) {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE pomodoro_sessions
		SET ended_at = ?
		WHERE id = ? AND ended_at IS NULL
	`, now, id)
	if err != nil {
		return nil, err
	}

	return s.GetSession(id)
}

// GetSessionStats 获取会话统计
func (s *Store) GetSessionStats(userID string, startTime, endTime time.Time) (*model.SessionStats, error) {
	var stats model.SessionStats
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as total_focus_minutes,
			COUNT(*) as total_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND started_at BETWEEN ? AND ? AND ended_at IS NOT NULL
	`, userID, startTime, endTime).Scan(&stats.TotalFocusMinutes, &stats.TotalSessions, &stats.CompletedTasks)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetDailyStats 获取每日统计
func (s *Store) GetDailyStats(userID string, startTime, endTime time.Time) ([]*model.DailyStats, error) {
	rows, err := s.db.Query(`
		SELECT 
			DATE(started_at) as date,
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as focus_minutes,
			COUNT(*) as completed_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND started_at BETWEEN ? AND ? AND ended_at IS NOT NULL
		GROUP BY DATE(started_at)
		ORDER BY date DESC
	`, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*model.DailyStats
	for rows.Next() {
		var stat model.DailyStats
		err := rows.Scan(&stat.Date, &stat.FocusMinutes, &stat.CompletedSessions, &stat.CompletedTasks)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
