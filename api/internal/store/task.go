package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

type TaskStore struct {
	db *sql.DB
}

func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{db: db}
}

// CreateTask 创建新任务
func (s *TaskStore) CreateTask(userID, title string, estimatedPomodoros int, color, notes string) (*model.Task, error) {
	task := &model.Task{
		ID:                 uuid.New().String(),
		UserID:             userID,
		Title:              title,
		EstimatedPomodoros: estimatedPomodoros,
		CompletedPomodoros: 0,
		Status:             model.TaskStatusPending,
		Color:              color,
		Notes:              notes,
		CreatedAt:          time.Now(),
	}

	query := `
		INSERT INTO tasks (id, user_id, title, estimated_pomodoros, completed_pomodoros, status, color, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, task.ID, task.UserID, task.Title, task.EstimatedPomodoros, task.CompletedPomodoros, task.Status, task.Color, task.Notes, task.CreatedAt)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTask 获取单个任务
func (s *TaskStore) GetTaskByID(id string) (*model.Task, error) {
	task := &model.Task{}
	query := `
		SELECT id, user_id, title, estimated_pomodoros, completed_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE id = ?
	`

	err := s.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.EstimatedPomodoros,
		&task.CompletedPomodoros,
		&task.Status,
		&task.CreatedAt,
		&task.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return task, nil
}

// ListTasks 获取用户的任务列表
func (s *TaskStore) GetUserTasks(userID string) ([]*model.Task, error) {
	query := `
		SELECT id, user_id, title, estimated_pomodoros, completed_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		task := &model.Task{}
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.EstimatedPomodoros,
			&task.CompletedPomodoros,
			&task.Status,
			&task.CreatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask 更新任务
func (s *TaskStore) UpdateTask(id string, req *model.TaskUpdate) (*model.Task, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	// 更新任务字段
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.EstimatedPomodoros != nil {
		task.EstimatedPomodoros = *req.EstimatedPomodoros
	}
	if req.Status != nil {
		task.Status = *req.Status
		if *req.Status == model.TaskStatusCompleted {
			now := time.Now()
			task.CompletedAt = &now
		}
	}

	// 保存更新
	query := `
		UPDATE tasks
		SET title = ?, estimated_pomodoros = ?, status = ?, completed_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query, task.Title, task.EstimatedPomodoros, task.Status, task.CompletedAt, task.ID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask 删除任务
func (s *TaskStore) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// GetCompletedTasksForDate 获取指定日期完成的任务（按完成时间倒序）
func (s *TaskStore) GetCompletedTasksForDate(userID string, date time.Time, limit int) ([]*model.TaskCompletion, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

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
		LIMIT ?
	`, userID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.TaskCompletion
	for rows.Next() {
		var task model.TaskCompletion
		var completedAt sql.NullTime
		err := rows.Scan(&task.TaskID, &task.TaskTitle, &task.Pomodoros, &task.TotalMinutes, &completedAt)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			task.CompletedAt = completedAt.Time
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskStore) UpdateTaskStatus(id string, status model.TaskStatus) error {
	var completedAt *time.Time
	if status == model.TaskStatusCompleted {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE tasks
		SET status = ?, completed_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, status, completedAt, id)
	return err
}

func (s *TaskStore) UpdateTaskProgress(id string, completedPomodoros int) error {
	query := `
		UPDATE tasks
		SET completed_pomodoros = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, completedPomodoros, id)
	return err
}
