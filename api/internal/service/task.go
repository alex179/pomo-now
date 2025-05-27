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

// ListTasks 获取用户的任务列表
func (s *TaskService) ListTasks(ctx context.Context, userID string) ([]*model.Task, error) {
	return s.store.ListTasks(userID)
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(ctx context.Context, userID string, req *model.TaskCreate) (*model.Task, error) {
	return s.store.CreateTask(userID, req)
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
func (s *TaskService) UpdateTask(ctx context.Context, id, userID string, req *model.TaskUpdate) (*model.Task, error) {
	task, err := s.store.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, apperrors.ErrTaskNotFound
	}
	if task.UserID != userID {
		return nil, apperrors.ErrNotAuthorized
	}

	return s.store.UpdateTask(id, req)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, id, userID string) error {
	task, err := s.store.GetTask(id)
	if err != nil {
		return err
	}
	if task == nil {
		return apperrors.ErrTaskNotFound
	}
	if task.UserID != userID {
		return apperrors.ErrNotAuthorized
	}

	return s.store.DeleteTask(id)
}
