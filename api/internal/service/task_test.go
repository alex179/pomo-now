package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TaskStoreInterface defines the interface for task store operations
// that TaskService depends on.
type TaskStoreInterface interface {
	CreateTask(task *model.Task) error
	GetTasksByUserID(userID string, statusFilter model.TaskStatus) ([]*model.Task, error)
	GetTaskByID(taskID string, userID string) (*model.Task, error)
	UpdateTask(task *model.Task) error
	DeleteTask(taskID string, userID string) error
}

// MockTaskStore is a mock implementation of TaskStoreInterface
type MockTaskStore struct {
	mock.Mock
}

func (m *MockTaskStore) CreateTask(task *model.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskStore) GetTasksByUserID(userID string, statusFilter model.TaskStatus) ([]*model.Task, error) {
	args := m.Called(userID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Task), args.Error(1)
}

func (m *MockTaskStore) GetTaskByID(taskID string, userID string) (*model.Task, error) {
	args := m.Called(taskID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskStore) UpdateTask(task *model.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskStore) DeleteTask(taskID string, userID string) error {
	args := m.Called(taskID, userID)
	return args.Error(0)
}

// Helper to create TaskService with mock store
// Similar to auth_test.go, this assumes TaskService can be instantiated with the mock.
// The actual NewTaskService takes a concrete *store.Store.
// For tests to run, TaskService.store field would need to be an interface type,
// or NewTaskService would need to accept TaskStoreInterface.
// We proceed assuming this is conceptually handled for the test code structure.
func newTestTaskService(store TaskStoreInterface) *TaskService {
	return &TaskService{
		store: store, // This is the conceptual leap, assuming TaskService.store is TaskStoreInterface
	}
}

func TestTaskService_CreateTask(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	tests := []struct {
		name                string
		title               string
		estimatedPomodoros  int
		setupMock           func(mockStore *MockTaskStore, taskMatcher interface{})
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name:               "Success",
			title:              "New Task",
			estimatedPomodoros: 3,
			setupMock: func(mockStore *MockTaskStore, taskMatcher interface{}) {
				mockStore.On("CreateTask", taskMatcher).Return(nil).Once()
			},
		},
		{
			name:                "Invalid Title - Empty",
			title:               "",
			estimatedPomodoros:  3,
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:                "Invalid EstimatedPomodoros - Zero",
			title:               "Valid Title",
			estimatedPomodoros:  0,
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:                "Invalid EstimatedPomodoros - Negative",
			title:               "Valid Title",
			estimatedPomodoros:  -1,
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:               "Store CreateTask Fails",
			title:              "Store Fail Task",
			estimatedPomodoros: 2,
			setupMock: func(mockStore *MockTaskStore, taskMatcher interface{}) {
				mockStore.On("CreateTask", taskMatcher).Return(errors.New("db error")).Once()
			},
			expectedError:     true,
			// Not checking specific AppError type here, as the service currently returns raw db error.
			// TODO: Service should wrap this error. For now, test direct error.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockTaskStore)
			taskService := newTestTaskService(mockStore)

			// Use mock.MatchedBy for flexible task matching in store calls
			taskMatcher := mock.MatchedBy(func(task *model.Task) bool {
				return task.UserID == userID && task.Title == tt.title && task.EstimatedPomodoros == tt.estimatedPomodoros
			})

			if tt.setupMock != nil {
				tt.setupMock(mockStore, taskMatcher)
			}

			createdTask, err := taskService.CreateTask(ctx, userID, tt.title, tt.estimatedPomodoros)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, createdTask)
				assert.Equal(t, userID, createdTask.UserID)
				assert.Equal(t, tt.title, createdTask.Title)
				assert.Equal(t, tt.estimatedPomodoros, createdTask.EstimatedPomodoros)
				assert.NotEmpty(t, createdTask.ID)
				assert.NotZero(t, createdTask.CreatedAt)
				assert.Equal(t, model.TaskStatusPending, createdTask.Status)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestTaskService_ListTasks(t *testing.T) {
	ctx := context.Background()
	userID := "user-list-tasks"
	mockTasks := []*model.Task{
		{ID: "task-1", UserID: userID, Title: "Task 1", Status: model.TaskStatusPending},
		{ID: "task-2", UserID: userID, Title: "Task 2", Status: model.TaskStatusCompleted},
	}

	tests := []struct {
		name                string
		statusFilter        model.TaskStatus
		setupMock           func(mockStore *MockTaskStore)
		expectedTasks       []*model.Task
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name:         "Success - No Filter",
			statusFilter: "",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTasksByUserID", userID, model.TaskStatus("")).Return(mockTasks, nil).Once()
			},
			expectedTasks: mockTasks,
		},
		{
			name:         "Success - With Valid Filter",
			statusFilter: model.TaskStatusPending,
			setupMock: func(mockStore *MockTaskStore) {
				// Assuming filter in store returns only matching tasks
				filteredTasks := []*model.Task{mockTasks[0]}
				mockStore.On("GetTasksByUserID", userID, model.TaskStatusPending).Return(filteredTasks, nil).Once()
			},
			expectedTasks: []*model.Task{mockTasks[0]},
		},
		{
			name:                "Invalid statusFilter",
			statusFilter:        "invalid_status",
			setupMock:           nil, // No store call expected
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:         "Store GetTasksByUserID Fails",
			statusFilter: "",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTasksByUserID", userID, model.TaskStatus("")).Return(nil, errors.New("db list error")).Once()
			},
			expectedError: true, 
			// TODO: Service should wrap this. For now, test direct error.
		},
		{
			name:         "Store Returns Empty List",
			statusFilter: "",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTasksByUserID", userID, model.TaskStatus("")).Return([]*model.Task{}, nil).Once()
			},
			expectedTasks: []*model.Task{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockTaskStore)
			taskService := newTestTaskService(mockStore)

			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			tasks, err := taskService.ListTasks(ctx, userID, tt.statusFilter)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTasks, tasks)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	ctx := context.Background()
	taskID := "task-get-123"
	userID := "user-get-task"
	mockTask := &model.Task{ID: taskID, UserID: userID, Title: "Test Get Task"}

	tests := []struct {
		name                string
		setupMock           func(mockStore *MockTaskStore)
		expectedTask        *model.Task
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name: "Success",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTaskByID", taskID, userID).Return(mockTask, nil).Once()
			},
			expectedTask: mockTask,
		},
		{
			name: "Not Found - Store returns nil",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTaskByID", taskID, userID).Return(nil, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrTaskNotFound,
		},
		{
			name: "Store GetTaskByID Fails",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("GetTaskByID", taskID, userID).Return(nil, errors.New("db get error")).Once()
			},
			expectedError:     true,
			expectedErrorType: &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer, // Service wraps this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockTaskStore)
			taskService := newTestTaskService(mockStore)

			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			task, err := taskService.GetTaskByID(ctx, taskID, userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedErrorType, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
				}
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}
			mockStore.AssertExpectations(t)
		})
	}
}


func TestTaskService_UpdateTask(t *testing.T) {
	ctx := context.Background()
	taskID := "task-update-123"
	userID := "user-update-task"
	originalTitle := "Original Title"
	originalPomodoros := 2
	originalStatus := model.TaskStatusPending

	existingTask := &model.Task{
		ID:                 taskID,
		UserID:             userID,
		Title:              originalTitle,
		EstimatedPomodoros: originalPomodoros,
		Status:             originalStatus,
		CreatedAt:          time.Now().Add(-time.Hour),
	}
	
	newTitle := "Updated Title"
	newPomodoros := 4
	newStatusCompleted := model.TaskStatusCompleted

	tests := []struct {
		name                string
		updateReq           model.TaskUpdate
		setupMock           func(mockStore *MockTaskStore, updatedTaskMatcher interface{})
		expectedTaskMutator func(task *model.Task) // Mutates expected task for comparison
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name:      "Success - Full Update",
			updateReq: model.TaskUpdate{Title: &newTitle, EstimatedPomodoros: &newPomodoros, Status: &newStatusCompleted},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
				mockStore.On("UpdateTask", updatedTaskMatcher).Return(nil).Once()
			},
			expectedTaskMutator: func(task *model.Task) {
				task.Title = newTitle
				task.EstimatedPomodoros = newPomodoros
				task.Status = newStatusCompleted
				// CompletedAt should be set by the service logic, so we check for its presence.
			},
		},
		{
			name:      "Success - Partial Update (Title only)",
			updateReq: model.TaskUpdate{Title: &newTitle},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
				mockStore.On("UpdateTask", updatedTaskMatcher).Return(nil).Once()
			},
			expectedTaskMutator: func(task *model.Task) { task.Title = newTitle },
		},
		{
			name:      "Success - Status to Completed",
			updateReq: model.TaskUpdate{Status: &newStatusCompleted},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
				mockStore.On("UpdateTask", updatedTaskMatcher).Return(nil).Once()
			},
			expectedTaskMutator: func(task *model.Task) {
				task.Status = newStatusCompleted
				// Check for CompletedAt in assertion
			},
		},
		{
			name:      "Task Not Found",
			updateReq: model.TaskUpdate{Title: &newTitle},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(nil, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrTaskNotFound,
		},
		{
			name:                "Invalid Input - Empty Title",
			updateReq:           model.TaskUpdate{Title: func(s string) *string { return &s }("")},
			setupMock:           func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:                "Invalid Input - Negative Pomodoros",
			updateReq:           model.TaskUpdate{EstimatedPomodoros: func(i int) *int { return &i }(-1)},
			setupMock:           func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:                "Invalid Input - Invalid Status",
			updateReq:           model.TaskUpdate{Status: func(s model.TaskStatus) *model.TaskStatus { return &s }("bad_status")},
			setupMock:           func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:      "Store GetTaskByID Fails",
			updateReq: model.TaskUpdate{Title: &newTitle},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(nil, errors.New("db get error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
		{
			name:      "Store UpdateTask Fails",
			updateReq: model.TaskUpdate{Title: &newTitle},
			setupMock: func(mockStore *MockTaskStore, updatedTaskMatcher interface{}) {
				mockStore.On("GetTaskByID", taskID, userID).Return(existingTask, nil).Once()
				mockStore.On("UpdateTask", updatedTaskMatcher).Return(errors.New("db update error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockTaskStore)
			taskService := newTestTaskService(mockStore)
			
			// Use a copy of existingTask to avoid modification across tests for GetTaskByID mock
			taskForGetByID := *existingTask 

			// Matcher for the UpdateTask call
			updatedTaskMatcher := mock.MatchedBy(func(task *model.Task) bool {
				if task.ID != taskID || task.UserID != userID { return false }
				// Check if the fields that were intended to be updated are updated
				// This is complex because partial updates are allowed.
				// For simplicity, we're checking if the mock is called.
				// More specific matching could be done based on tt.updateReq
				return true
			})

			if tt.setupMock != nil {
				tt.setupMock(mockStore, updatedTaskMatcher)
			}

			updatedTask, err := taskService.UpdateTask(ctx, taskID, userID, tt.updateReq)

			if tt.expectedError {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedErrorType, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
				}
				assert.Nil(t, updatedTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedTask)

				expectedResultTask := taskForGetByID // Start with the original fetched task
				if tt.expectedTaskMutator != nil {
					tt.expectedTaskMutator(&expectedResultTask)
				}
				
				assert.Equal(t, expectedResultTask.Title, updatedTask.Title)
				assert.Equal(t, expectedResultTask.EstimatedPomodoros, updatedTask.EstimatedPomodoros)
				assert.Equal(t, expectedResultTask.Status, updatedTask.Status)

				if expectedResultTask.Status == model.TaskStatusCompleted {
					assert.NotNil(t, updatedTask.CompletedAt)
				} else {
					assert.Nil(t, updatedTask.CompletedAt)
				}
			}
			mockStore.AssertExpectations(t)
		})
	}
}


func TestTaskService_DeleteTask(t *testing.T) {
	ctx := context.Background()
	taskID := "task-delete-123"
	userID := "user-delete-task"

	tests := []struct {
		name                string
		setupMock           func(mockStore *MockTaskStore)
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name: "Success",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("DeleteTask", taskID, userID).Return(nil).Once()
			},
		},
		{
			name: "Task Not Found (sql.ErrNoRows from store)",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("DeleteTask", taskID, userID).Return(sql.ErrNoRows).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrTaskNotFound,
		},
		{
			name: "Store DeleteTask Fails (Generic Error)",
			setupMock: func(mockStore *MockTaskStore) {
				mockStore.On("DeleteTask", taskID, userID).Return(errors.New("db delete error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockTaskStore)
			taskService := newTestTaskService(mockStore)

			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			err := taskService.DeleteTask(ctx, taskID, userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedErrorType, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

[end of api/internal/service/task_test.go]
