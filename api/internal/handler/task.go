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

// Tasks 处理任务列表和创建请求
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

// Task 处理单个任务的获取、更新和删除请求
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
	// Get status query parameter
	statusFilter := r.URL.Query().Get("status")

	tasks, err := h.taskService.ListTasks(r.Context(), user.ID, model.TaskStatus(statusFilter))
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrBadRequest:
				response.Error(w, http.StatusBadRequest, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, tasks)
}

// createTask 创建新任务
func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request, user *model.User) {
	var req model.TaskCreate // This struct contains Title and EstimatedPomodoros
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate incoming data (though service layer also validates)
	// Basic validation can happen here too, e.g. presence of required fields
	if req.Title == "" {
		response.Error(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.EstimatedPomodoros <= 0 {
		// Assuming EstimatedPomodoros is required and must be positive.
		// The model.TaskCreate has validate tags, but direct check here is also fine.
		response.Error(w, http.StatusBadRequest, "estimated_pomodoros must be positive")
		return
	}


	task, err := h.taskService.CreateTask(r.Context(), user.ID, req.Title, req.EstimatedPomodoros)
	if err != nil {
		// Determine appropriate HTTP status code based on the error type from service
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrBadRequest:
				response.Error(w, http.StatusBadRequest, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
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

	// Ensure at least one field is provided for update
	if req.Title == nil && req.EstimatedPomodoros == nil && req.Status == nil {
		response.Error(w, http.StatusBadRequest, "at least one field must be provided for update")
		return
	}


	task, err := h.taskService.UpdateTask(r.Context(), taskID, user.ID, req) // Pass req by value as service expects model.TaskUpdate
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrTaskNotFound:
				response.Error(w, http.StatusNotFound, appErr.Message)
			case apperrors.ErrNotAuthorized:
				response.Error(w, http.StatusForbidden, appErr.Message)
			case apperrors.ErrBadRequest:
				response.Error(w, http.StatusBadRequest, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, task)
}

// deleteTask 删除任务
func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, user *model.User, taskID string) {
	err := h.taskService.DeleteTask(r.Context(), taskID, user.ID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.Code {
			case apperrors.ErrTaskNotFound:
				response.Error(w, http.StatusNotFound, appErr.Message)
			case apperrors.ErrNotAuthorized: // This might be covered by TaskNotFound if DeleteTask doesn't distinguish
				response.Error(w, http.StatusForbidden, appErr.Message)
			default:
				response.Error(w, http.StatusInternalServerError, appErr.Error())
			}
		} else {
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.NoContent(w)
}
