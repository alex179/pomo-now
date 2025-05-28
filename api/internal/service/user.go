package service

import (
	"context"
	// "database/sql" // Not directly needed here, but store might return sql.ErrNoRows

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
)

// UserService handles business logic related to user preferences and profiles.
type UserService struct {
	store *store.Store // Using concrete store type as per existing services
}

// NewUserService creates a new UserService.
func NewUserService(store *store.Store) *UserService {
	return &UserService{store: store}
}

// GetPreferences retrieves a user's preferences.
// Returns a model.User containing ID and preference fields.
func (s *UserService) GetPreferences(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		// Consider if sql.ErrNoRows should be mapped to apperrors.ErrUserNotFound
		// For now, assume GetUserByID might return a wrapped error or nil,nil for not found.
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to retrieve user", err)
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}

	// Return only relevant fields for preferences to avoid exposing sensitive data if not needed by caller
	// However, the task allows returning the full user object. For simplicity and consistency with GetUserByID,
	// we return the user object which now includes these fields.
	// If we wanted to return a specific preferences-only struct, we'd map it here.
	return user, nil
}

// UpdatePreferences updates a user's preferences.
func (s *UserService) UpdatePreferences(ctx context.Context, userID string, req model.UpdateUserPreferencesRequest) (*model.User, error) {
	// Validate input values
	if req.WorkDuration != nil && *req.WorkDuration <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "work_duration must be positive", nil)
	}
	if req.ShortBreakDuration != nil && *req.ShortBreakDuration <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "short_break_duration must be positive", nil)
	}
	if req.LongBreakDuration != nil && *req.LongBreakDuration <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "long_break_duration must be positive", nil)
	}
	if req.LongBreakInterval != nil && *req.LongBreakInterval <= 0 {
		return nil, apperrors.NewAppError(apperrors.ErrBadRequest, "long_break_interval must be positive", nil)
	}

	// Create a model.User object with only the preference fields set from the request
	userWithPrefsOnly := &model.User{
		WorkDuration:       req.WorkDuration,
		ShortBreakDuration: req.ShortBreakDuration,
		LongBreakDuration:  req.LongBreakDuration,
		LongBreakInterval:  req.LongBreakInterval,
	}

	err := s.store.UpdateUserPreferences(userID, userWithPrefsOnly)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to update preferences", err)
	}

	// After successful update, get the updated user object
	updatedUser, err := s.store.GetUserByID(userID)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.ErrInternalServer, "failed to retrieve updated user preferences", err)
	}
	if updatedUser == nil {
		// This case should ideally not happen if update succeeded and user exists.
		return nil, apperrors.ErrUserNotFound 
	}

	return updatedUser, nil
}
