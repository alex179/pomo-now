package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

// CreateTask 创建新任务
func (s *Store) CreateTask(userID string, req *model.TaskCreate) (*model.Task, error) {
	task := &model.Task{
		ID:                 uuid.New().String(),
		UserID:             userID,
		Title:              req.Title,
		EstimatedPomodoros: req.EstimatedPomodoros,
		Status:             model.TaskStatusPending,
		CreatedAt:          time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO tasks (id, user_id, title, estimated_pomodoros, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, task.ID, task.UserID, task.Title, task.EstimatedPomodoros, task.Status, task.CreatedAt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask 获取单个任务
func (s *Store) GetTask(id string) (*model.Task, error) {
	var task model.Task
	var completedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, user_id, title, estimated_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(&task.ID, &task.UserID, &task.Title, &task.EstimatedPomodoros, &task.Status, &task.CreatedAt, &completedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	return &task, nil
}

// ListTasks 获取用户的任务列表
func (s *Store) ListTasks(userID string) ([]*model.Task, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, title, estimated_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		var completedAt sql.NullTime
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.EstimatedPomodoros, &task.Status, &task.CreatedAt, &completedAt)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask 更新任务
func (s *Store) UpdateTask(id string, req *model.TaskUpdate) (*model.Task, error) {
	task, err := s.GetTask(id)
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
	_, err = s.db.Exec(`
		UPDATE tasks
		SET title = ?, estimated_pomodoros = ?, status = ?, completed_at = ?
		WHERE id = ?
	`, task.Title, task.EstimatedPomodoros, task.Status, task.CompletedAt, task.ID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask 删除任务
func (s *Store) DeleteTask(id string) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}
