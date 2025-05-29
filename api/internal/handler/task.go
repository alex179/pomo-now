package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alex/pomo-now/internal/middleware"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/response"
	"github.com/alex/pomo-now/internal/service"
)

// TaskHandler 处理任务相关的请求
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// Tasks 处理任务列表请求
func (h *TaskHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r, user)
	case http.MethodPost:
		h.createTask(w, r, user)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Task 处理单个任务请求
func (h *TaskHandler) Task(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := r.PathValue("id")
	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "task id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, user, taskID)
	case http.MethodPut:
		h.updateTask(w, r, user, taskID)
	case http.MethodDelete:
		h.deleteTask(w, r, user, taskID)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listTasks 获取任务列表
func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request, user *model.User) {
	tasks, err := h.taskService.ListTasks(r.Context(), user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, tasks)
}

// createTask 创建新任务
func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request, user *model.User) {
	var req model.TaskCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), user.ID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(w, task)
}

// getTask 获取单个任务
func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request, user *model.User, taskID string) {
	task, err := h.taskService.GetTask(r.Context(), taskID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		response.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if task.UserID != user.ID {
		response.Error(w, http.StatusForbidden, "not authorized to access this task")
		return
	}

	response.Success(w, task)
}

// updateTask 更新任务
func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request, user *model.User, taskID string) {
	var req model.TaskUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), taskID, user.ID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, task)
}

// deleteTask 删除任务
func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, user *model.User, taskID string) {
	if err := h.taskService.DeleteTask(r.Context(), taskID, user.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// HandleTasks 处理 /api/tasks 路由
func (h *TaskHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	h.Tasks(w, r)
}

// HandleTaskByID 处理 /api/tasks/{id} 路由
func (h *TaskHandler) HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	h.Task(w, r)
}

// GetTasks 公共方法用于路由
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.listTasks(w, r, user)
}

// CreateTask 公共方法用于路由
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.createTask(w, r, user)
}

// GetTask 公共方法用于路由
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "task id is required")
		return
	}
	h.getTask(w, r, user, taskID)
}

// UpdateTask 公共方法用于路由
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "task id is required")
		return
	}
	h.updateTask(w, r, user, taskID)
}

// DeleteTask 公共方法用于路由
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := r.PathValue("id")
	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "task id is required")
		return
	}
	h.deleteTask(w, r, user, taskID)
}
