package store

import (
	"database/sql"
	"testing"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	// TODO: Replace with test database setup
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			estimated_pomodoros INTEGER NOT NULL,
			completed_pomodoros INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			color TEXT,
			notes TEXT,
			created_at TIMESTAMP NOT NULL,
			completed_at TIMESTAMP
		)
	`)
	require.NoError(t, err)

	return db
}

func TestTaskStore_CreateTask(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	userID := uuid.New().String()
	title := "Test Task"
	estimatedPomodoros := 4
	color := "#FF0000"
	notes := "Test notes"

	task, err := store.CreateTask(userID, title, estimatedPomodoros, color, notes)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, userID, task.UserID)
	assert.Equal(t, title, task.Title)
	assert.Equal(t, estimatedPomodoros, task.EstimatedPomodoros)
	assert.Equal(t, color, task.Color)
	assert.Equal(t, notes, task.Notes)
	assert.Equal(t, model.TaskStatusPending, task.Status)
	assert.Equal(t, 0, task.CompletedPomodoros)
}

func TestTaskStore_GetTaskByID(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	// Create a task first
	userID := uuid.New().String()
	task, err := store.CreateTask(userID, "Test Task", 4, "#FF0000", "Test notes")
	require.NoError(t, err)

	// Test getting the task
	retrievedTask, err := store.GetTaskByID(task.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedTask)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.Title, retrievedTask.Title)
}

func TestTaskStore_UpdateTask(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	// Create a task first
	userID := uuid.New().String()
	task, err := store.CreateTask(userID, "Test Task", 4, "#FF0000", "Test notes")
	require.NoError(t, err)

	// Update the task
	newTitle := "Updated Task"
	newEstimatedPomodoros := 6
	update := &model.TaskUpdate{
		Title:              &newTitle,
		EstimatedPomodoros: &newEstimatedPomodoros,
	}

	updatedTask, err := store.UpdateTask(task.ID, update)
	require.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, newTitle, updatedTask.Title)
	assert.Equal(t, newEstimatedPomodoros, updatedTask.EstimatedPomodoros)
}

func TestTaskStore_DeleteTask(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	// Create a task first
	userID := uuid.New().String()
	task, err := store.CreateTask(userID, "Test Task", 4, "#FF0000", "Test notes")
	require.NoError(t, err)

	// Delete the task
	err = store.DeleteTask(task.ID)
	require.NoError(t, err)

	// Verify task is deleted
	_, err = store.GetTaskByID(task.ID)
	assert.Error(t, err)
}

func TestTaskStore_UpdateTaskStatus(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	// Create a task first
	userID := uuid.New().String()
	task, err := store.CreateTask(userID, "Test Task", 4, "#FF0000", "Test notes")
	require.NoError(t, err)

	// Update status to completed
	err = store.UpdateTaskStatus(task.ID, model.TaskStatusCompleted)
	require.NoError(t, err)

	// Verify status is updated
	updatedTask, err := store.GetTaskByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusCompleted, updatedTask.Status)
	assert.NotNil(t, updatedTask.CompletedAt)
}

func TestTaskStore_UpdateTaskProgress(t *testing.T) {
	db := setupTestDB(t)
	store := NewTaskStore(db)

	// Create a task first
	userID := uuid.New().String()
	task, err := store.CreateTask(userID, "Test Task", 4, "#FF0000", "Test notes")
	require.NoError(t, err)

	// Update progress
	completedPomodoros := 2
	err = store.UpdateTaskProgress(task.ID, completedPomodoros)
	require.NoError(t, err)

	// Verify progress is updated
	updatedTask, err := store.GetTaskByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, completedPomodoros, updatedTask.CompletedPomodoros)
}
