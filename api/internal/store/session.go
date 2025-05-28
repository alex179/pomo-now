package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

// CreateSession inserts a new session record into the database.
// It expects session.ID and session.EndedAt to be set by the service.
func (s *Store) CreateSession(session *model.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	// EndedAt is expected to be set by the service layer.

	_, err := s.db.Exec(`
		INSERT INTO pomodoro_sessions (id, user_id, task_id, duration_minutes, type, ended_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.TaskID, session.DurationMinutes, session.Type, session.EndedAt)
	
	return err
}

// GetSession 获取单个会话
func (s *Store) GetSession(id string) (*model.Session, error) {
	var session model.Session
	// var endedAt sql.NullTime // EndedAt is now non-nullable in model.Session
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE id = ?
	`, id).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// if endedAt.Valid { // No longer needed as EndedAt is time.Time
	// 	session.EndedAt = endedAt.Time
	// }
	return &session, nil
}

// GetCurrentSession 获取用户的当前会话
// Note: The concept of "current session" (ended_at IS NULL) is removed due to schema change (ended_at default current_timestamp)
// This function might need re-evaluation or removal based on new app logic.
// For now, let's assume it might be used to fetch the *last recorded* session if needed,
// but its original purpose (a session that hasn't ended) is gone.
// To keep it simple, I will comment it out for now as it relies on ended_at IS NULL which is no longer the primary way to identify an active session.
/*
func (s *Store) GetCurrentSession(userID string) (*model.Session, error) {
	var session model.Session
	// var endedAt sql.NullTime // EndedAt is now non-nullable
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? -- AND ended_at IS NULL (This condition is problematic with new schema)
		ORDER BY ended_at DESC -- Order by ended_at to get the latest one
		LIMIT 1
	`, userID).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}
*/

// ListSessions 获取用户的会话列表
func (s *Store) ListSessions(userID string, startTime, endTime time.Time) ([]*model.Session, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ? -- Query by ended_at
		ORDER BY ended_at DESC
	`, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var session model.Session
		// var endedAt sql.NullTime // EndedAt is now non-nullable
		err := rows.Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
		if err != nil {
			return nil, err
		}
		// if endedAt.Valid {
		// 	session.EndedAt = &endedAt.Time
		// }
		sessions = append(sessions, &session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// EndSession 结束会话
// This function is likely obsolete as EndedAt is now set at creation time.
// If a session's end time needs to be updated after creation, a different mechanism or UpdateSession would be needed.
// For now, commenting out as its original purpose is fulfilled by CreateSession.
/*
func (s *Store) EndSession(id string) (*model.Session, error) {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE pomodoro_sessions
		SET ended_at = ?
		WHERE id = ? -- AND ended_at IS NULL (original logic, might not apply)
	`, now, id)
	if err != nil {
		return nil, err
	}

	return s.GetSession(id)
}
*/

// GetSessionStats 获取会话统计
func (s *Store) GetSessionStats(userID string, startTime, endTime time.Time) (*model.SessionStats, error) {
	var stats model.SessionStats
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as total_focus_minutes,
			COUNT(*) as total_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ? -- Query by ended_at
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
			DATE(ended_at) as date, -- Group by DATE(ended_at)
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as focus_minutes,
			COUNT(*) as completed_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ?
		GROUP BY DATE(ended_at)
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
	}
	return &session, nil
}

	var session model.Session
	// var endedAt sql.NullTime // EndedAt is now non-nullable in model.Session
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE id = ?
	`, id).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// if endedAt.Valid { // No longer needed as EndedAt is time.Time
	// 	session.EndedAt = endedAt.Time
	// }
	return &session, nil
}

// GetCurrentSession 获取用户的当前会话
// Note: The concept of "current session" (ended_at IS NULL) is removed due to schema change (ended_at default current_timestamp)
// This function might need re-evaluation or removal based on new app logic.
// For now, let's assume it might be used to fetch the *last recorded* session if needed,
// but its original purpose (a session that hasn't ended) is gone.
// To keep it simple, I will comment it out for now as it relies on ended_at IS NULL which is no longer the primary way to identify an active session.
/*
func (s *Store) GetCurrentSession(userID string) (*model.Session, error) {
	var session model.Session
	// var endedAt sql.NullTime // EndedAt is now non-nullable
	err := s.db.QueryRow(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? -- AND ended_at IS NULL (This condition is problematic with new schema)
		ORDER BY ended_at DESC -- Order by ended_at to get the latest one
		LIMIT 1
	`, userID).Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}
*/

// ListSessions 获取用户的会话列表
func (s *Store) ListSessions(userID string, startTime, endTime time.Time) ([]*model.Session, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, task_id, duration_minutes, type, ended_at
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ? -- Query by ended_at
		ORDER BY ended_at DESC
	`, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var session model.Session
		// var endedAt sql.NullTime // EndedAt is now non-nullable
		err := rows.Scan(&session.ID, &session.UserID, &session.TaskID, &session.DurationMinutes, &session.Type, &session.EndedAt)
		if err != nil {
			return nil, err
		}
		// if endedAt.Valid {
		// 	session.EndedAt = &endedAt.Time
		// }
		sessions = append(sessions, &session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// EndSession 结束会话
// This function is likely obsolete as EndedAt is now set at creation time.
// If a session's end time needs to be updated after creation, a different mechanism or UpdateSession would be needed.
// For now, commenting out as its original purpose is fulfilled by CreateSession.
/*
func (s *Store) EndSession(id string) (*model.Session, error) {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE pomodoro_sessions
		SET ended_at = ?
		WHERE id = ? -- AND ended_at IS NULL (original logic, might not apply)
	`, now, id)
	if err != nil {
		return nil, err
	}

	return s.GetSession(id)
}
*/

// GetSessionStats 获取会话统计
func (s *Store) GetSessionStats(userID string, startTime, endTime time.Time) (*model.SessionStats, error) {
	var stats model.SessionStats
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as total_focus_minutes,
			COUNT(*) as total_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ? -- Query by ended_at
	`, userID, startTime, endTime).Scan(&stats.TotalFocusMinutes, &stats.TotalSessions, &stats.CompletedTasks)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetPomodoroStatsByUserID retrieves total focus sessions and minutes for a user within a period.
func (s *Store) GetPomodoroStatsByUserID(userID string, startTime time.Time, endTime time.Time) (totalFocusSessions int, totalFocusMinutes int, err error) {
	row := s.db.QueryRow(`
		SELECT
			COALESCE(COUNT(*), 0),
			COALESCE(SUM(duration_minutes), 0)
		FROM pomodoro_sessions
		WHERE user_id = ? AND type = ? AND ended_at BETWEEN ? AND ?
	`, userID, model.SessionTypeFocus, startTime, endTime)

	err = row.Scan(&totalFocusSessions, &totalFocusMinutes)
	if err != nil {
		// If no rows are found, Scan might return sql.ErrNoRows.
		// In such a case, counts should be 0, which COALESCE handles in SQL.
		// If Scan returns an error other than sql.ErrNoRows, it's a genuine error.
		if err == sql.ErrNoRows {
			return 0, 0, nil // No focus sessions found, not an error.
		}
		return 0, 0, err
	}
	return totalFocusSessions, totalFocusMinutes, nil
}

// GetCompletedTasksCountByUserID retrieves the count of tasks completed by a user within a period.
func (s *Store) GetCompletedTasksCountByUserID(userID string, startTime time.Time, endTime time.Time) (int, error) {
	var count int
	// Assuming 'tasks' table has 'user_id', 'status', and 'completed_at' columns.
	// And model.TaskStatusCompleted is the value for completed tasks.
	err := s.db.QueryRow(`
		SELECT COALESCE(COUNT(*), 0)
		FROM tasks
		WHERE user_id = ? AND status = ? AND completed_at BETWEEN ? AND ?
	`, userID, model.TaskStatusCompleted, startTime, endTime).Scan(&count)

	if err != nil {
		if err == sql.ErrNoRows { // Should not happen with COUNT, but good practice
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// GetDailyStats 获取每日统计
func (s *Store) GetDailyStats(userID string, startTime, endTime time.Time) ([]*model.DailyStats, error) {
	rows, err := s.db.Query(`
		SELECT 
			DATE(ended_at) as date, -- Group by DATE(ended_at)
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as focus_minutes,
			COUNT(*) as completed_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND ended_at BETWEEN ? AND ?
		GROUP BY DATE(ended_at)
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
