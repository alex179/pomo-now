package store

import (
	"database/sql"
	"fmt"
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
			COALESCE(COUNT(CASE WHEN type = 'focus' AND ended_at IS NOT NULL THEN 1 END), 0) as total_pomodoros,
			COALESCE(SUM(CASE WHEN ended_at IS NOT NULL THEN duration_minutes ELSE 0 END), 0) as total_minutes,
			COALESCE(SUM(CASE WHEN type = 'focus' THEN duration_minutes ELSE 0 END), 0) as total_focus_minutes,
			COUNT(*) as total_sessions,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND started_at BETWEEN ? AND ? AND ended_at IS NOT NULL
	`, userID, startTime, endTime).Scan(&stats.TotalPomodoros, &stats.TotalMinutes, &stats.TotalFocusMinutes, &stats.TotalSessions, &stats.CompletedTasks)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetCompletedTasksCount 获取指定时间范围内完成的任务数量
func (s *Store) GetCompletedTasksCount(userID string, startTime, endTime time.Time) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = ? AND status = 'completed' AND completed_at BETWEEN ? AND ?
	`, userID, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTasksCount 获取指定时间范围内的任务总数（用于计算完成率）
func (s *Store) GetTasksCount(userID string, startTime, endTime time.Time) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = ? AND created_at BETWEEN ? AND ?
	`, userID, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
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

// GetHourlyStats 获取指定日期的每小时统计
func (s *Store) GetHourlyStats(userID string, date time.Time) ([]*model.HourlyStats, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	rows, err := s.db.Query(`
		SELECT 
			HOUR(started_at) as hour,
			COUNT(*) as sessions,
			COALESCE(SUM(CASE WHEN ended_at IS NOT NULL THEN duration_minutes ELSE 0 END), 0) as minutes
		FROM pomodoro_sessions
		WHERE user_id = ? AND type = 'focus' AND started_at BETWEEN ? AND ?
		GROUP BY HOUR(started_at)
		ORDER BY hour
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 创建24小时的完整统计，初始化为0
	hourlyStats := make([]*model.HourlyStats, 24)
	for i := 0; i < 24; i++ {
		hourlyStats[i] = &model.HourlyStats{
			Hour:       i,
			HourStr:    fmt.Sprintf("%02d:00", i),
			Sessions:   0,
			Minutes:    0,
			Percentage: 0,
		}
	}

	// 计算总分钟数用于百分比计算
	var totalMinutes int
	for rows.Next() {
		var hour, sessions, minutes int
		err := rows.Scan(&hour, &sessions, &minutes)
		if err != nil {
			return nil, err
		}
		hourlyStats[hour].Sessions = sessions
		hourlyStats[hour].Minutes = minutes
		totalMinutes += minutes
	}

	// 计算百分比
	for _, stat := range hourlyStats {
		if totalMinutes > 0 {
			stat.Percentage = float64(stat.Minutes) / float64(totalMinutes) * 100
		}
	}

	return hourlyStats, nil
}

// GetDayDetail 获取特定日期的详细统计
func (s *Store) GetDayDetail(userID string, date time.Time) (*model.DayDetail, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	// 获取基本统计
	var detail model.DayDetail
	detail.Date = start

	err := s.db.QueryRow(`
		SELECT 
			COALESCE(COUNT(CASE WHEN type = 'focus' AND ended_at IS NOT NULL THEN 1 END), 0) as total_pomodoros,
			COALESCE(SUM(CASE WHEN ended_at IS NOT NULL THEN duration_minutes ELSE 0 END), 0) as total_minutes,
			COUNT(DISTINCT task_id) as completed_tasks
		FROM pomodoro_sessions
		WHERE user_id = ? AND started_at BETWEEN ? AND ?
	`, userID, start, end).Scan(&detail.TotalPomodoros, &detail.TotalMinutes, &detail.CompletedTasks)
	if err != nil {
		return nil, err
	}

	// 获取每小时统计
	hourlyStats, err := s.GetHourlyStats(userID, date)
	if err != nil {
		return nil, err
	}
	detail.HourlyStats = hourlyStats

	// 获取任务完成记录
	taskDetails, err := s.GetTaskCompletionDetails(userID, start, end)
	if err != nil {
		return nil, err
	}
	detail.TaskDetails = taskDetails

	return &detail, nil
}

// GetTaskCompletionDetails 获取指定时间范围内的任务完成详情
func (s *Store) GetTaskCompletionDetails(userID string, start, end time.Time) ([]*model.TaskCompletion, error) {
	rows, err := s.db.Query(`
		SELECT 
			t.id,
			t.title,
			COUNT(ps.id) as pomodoros,
			COALESCE(SUM(ps.duration_minutes), 0) as total_minutes,
			t.completed_at
		FROM tasks t
		LEFT JOIN pomodoro_sessions ps ON t.id = ps.task_id AND ps.type = 'focus' AND ps.ended_at IS NOT NULL
		WHERE t.user_id = ? AND t.status = 'completed' AND t.completed_at BETWEEN ? AND ?
		GROUP BY t.id, t.title, t.completed_at
		ORDER BY t.completed_at DESC
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []*model.TaskCompletion
	for rows.Next() {
		var detail model.TaskCompletion
		var completedAt sql.NullTime
		err := rows.Scan(&detail.TaskID, &detail.TaskTitle, &detail.Pomodoros, &detail.TotalMinutes, &completedAt)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			detail.CompletedAt = completedAt.Time
		}
		details = append(details, &detail)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return details, nil
}
