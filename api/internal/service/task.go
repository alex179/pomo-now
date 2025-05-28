package service

import (
	"context"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
)

// TaskService 处理任务相关的业务逻辑
type TaskService struct {
	store *store.Store
}

// NewTaskService 创建新的任务服务
func NewTaskService(store *store.Store) *TaskService {
	return &TaskService{
		store: store,
	}
}

// ListTasks retrieves a list of tasks for a given user, optionally filtered by status.
func (s *TaskService) ListTasks(ctx context.Context, userID string, statusFilter model.TaskStatus) ([]*model.Task, error) {
	// Validate statusFilter if necessary (e.g., ensure it's one of the defined TaskStatus values)
	// For now, we assume any string can be passed, and the store handles empty string for no filter.
	// If statusFilter is not empty, it should be one of model.TaskStatusPending, model.TaskStatusInProgress, etc.
	if statusFilter != "" {
		isValidStatus := false
		for _, validStatus := range []model.TaskStatus{model.TaskStatusPending, model.TaskStatusInProgress, model.TaskStatusCompleted} {
			if statusFilter == validStatus {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "invalid status filter value", nil)
		}
	}
	return s.store.GetTasksByUserID(userID, statusFilter)
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(ctx context.Context, userID string, title string, estimatedPomodoros int) (*model.Task, error) {
	// Validate input
	if title == "" {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "title cannot be empty", nil)
	}
	if estimatedPomodoros <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "estimated_pomodoros must be greater than 0", nil)
	}

	task := &model.Task{
		// ID and CreatedAt will be set by the store
		UserID:             userID,
		Title:              title,
		EstimatedPomodoros: estimatedPomodoros,
		Status:             model.TaskStatusPending, // Default status
	}

	err := s.store.CreateTask(task)
	if err != nil {
		// TODO: Wrap error for more context? e.g., apperrors.NewAppError(apperrors.ErrInternalServer, "failed to create task", err)
		return nil, err
	}

	return task, nil
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	task, err := s.store.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, apperrors.ErrTaskNotFound
	}
	return task, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(ctx context.Context, taskID string, userID string, updateReq model.TaskUpdate) (*model.Task, error) {
	// Fetch the existing task, ensuring it belongs to the user
	task, err := s.store.GetTaskByID(taskID, userID)
	if err != nil {
		// Distinguish between DB error and not found potentially, though GetTaskByID might do that
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to retrieve task for update", err)
	}
	if task == nil {
		return nil, apperrors.ErrTaskNotFound // Or ErrNotAuthorized if GetTaskByID implies that by returning nil
	}

	// Apply changes from updateReq
	if updateReq.Title != nil {
		if *updateReq.Title == "" {
			return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "title cannot be empty", nil)
		}
		task.Title = *updateReq.Title
	}
	if updateReq.EstimatedPomodoros != nil {
		if *updateReq.EstimatedPomodoros <= 0 {
			return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "estimated_pomodoros must be greater than 0", nil)
		}
		task.EstimatedPomodoros = *updateReq.EstimatedPomodoros
	}
	if updateReq.Status != nil {
		// Validate new status value
		isValidStatus := false
		for _, validStatus := range []model.TaskStatus{model.TaskStatusPending, model.TaskStatusInProgress, model.TaskStatusCompleted} {
			if *updateReq.Status == validStatus {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "invalid status value", nil)
		}
		task.Status = *updateReq.Status

		// Handle CompletedAt based on Status
		if task.Status == model.TaskStatusCompleted {
			if task.CompletedAt == nil { // Only set if not already completed (idempotency)
				now := time.Now().UTC()
				task.CompletedAt = &now
			}
		} else {
			task.CompletedAt = nil // Set to null if status is not completed
		}
	}

	// Call store.UpdateTask with the modified task object
	err = s.store.UpdateTask(task)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to update task", err)
	}

	return task, nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID string, userID string) error {
	// The store's DeleteTask now includes userID check, so pre-fetching is not strictly necessary
	// for authorization, but can be useful for a more specific "not found" vs "not authorized".
	// However, store.DeleteTask returning sql.ErrNoRows if no row matched (id AND user_id) is sufficient.
	err := s.store.DeleteTask(taskID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Assuming store.DeleteTask returns sql.ErrNoRows
			return apperrors.ErrTaskNotFound // Or ErrNotAuthorized, as the task for that user wasn't found
		}
		return apperrors.NewAppError(apperrors.ErrInternalServer, "failed to delete task", err)
	}
	return nil
}
