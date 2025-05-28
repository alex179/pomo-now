package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
)

// CreateTask inserts a new task record into the database.
// It sets the ID and CreatedAt fields if they are not already set.
func (s *Store) CreateTask(task *model.Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC() // Use UTC for consistency
	}
	// Ensure status is set if not provided, defaulting to pending
	if task.Status == "" {
		task.Status = model.TaskStatusPending
	}


	_, err := s.db.Exec(`
		INSERT INTO tasks (id, user_id, title, estimated_pomodoros, status, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.UserID, task.Title, task.EstimatedPomodoros, task.Status, task.CreatedAt, task.CompletedAt)
	// Note: completed_at is included in the INSERT, will be NULL if task.CompletedAt is nil.
	
	return err
}

// GetTaskByID retrieves a specific task by its ID and ensures it belongs to the given userID.
func (s *Store) GetTaskByID(taskID string, userID string) (*model.Task, error) {
	var task model.Task
	var completedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, user_id, title, estimated_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE id = ? AND user_id = ?
	`, taskID, userID).Scan(&task.ID, &task.UserID, &task.Title, &task.EstimatedPomodoros, &task.Status, &task.CreatedAt, &completedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Task not found for this user or does not exist
	}
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	return &task, nil
}

// GetTasksByUserID retrieves a list of tasks for a given user, optionally filtered by status.
func (s *Store) GetTasksByUserID(userID string, statusFilter model.TaskStatus) ([]*model.Task, error) {
	query := `
		SELECT id, user_id, title, estimated_pomodoros, status, created_at, completed_at
		FROM tasks
		WHERE user_id = ?
	`
	args := []interface{}{userID}

	if statusFilter != "" {
		query += " AND status = ?"
		args = append(args, statusFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
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
			return nil, err // Error during row scan
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Error after iterating rows
	}

	return tasks, nil
}

// UpdateTask updates an existing task in the database.
// It assumes the task object passed in contains the new values and the correct ID.
// UserID check should happen in the service layer by first fetching the task via GetTaskByID.
func (s *Store) UpdateTask(task *model.Task) error {
	// Ensure CompletedAt is handled correctly based on Status
	if task.Status == model.TaskStatusCompleted && task.CompletedAt == nil {
		now := time.Now().UTC()
		task.CompletedAt = &now
	} else if task.Status != model.TaskStatusCompleted {
		task.CompletedAt = nil // Set to null if not completed
	}

	_, err := s.db.Exec(`
		UPDATE tasks
		SET title = ?, estimated_pomodoros = ?, status = ?, completed_at = ?
		WHERE id = ? 
	`, task.Title, task.EstimatedPomodoros, task.Status, task.CompletedAt, task.ID)
	// Note: We are not updating user_id or created_at
	return err
}

// DeleteTask deletes a task by its ID, ensuring it belongs to the userID.
func (s *Store) DeleteTask(taskID string, userID string) error {
	res, err := s.db.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Or a custom error indicating task not found for user or already deleted
	}
	return nil
}
