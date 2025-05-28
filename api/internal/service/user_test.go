package service

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// UserStoreInterface defines the interface for user store operations
// that UserService and AuthService depend on.
type UserStoreInterface interface {
	GetUserByID(id string) (*model.User, error)
	UpdateUserPreferences(userID string, preferences *model.User) error
	// Methods needed by AuthService (from previous auth_test.go)
	GetUserByEmail(email string) (*model.User, error)
	CreateEmailUser(email, password string) (*model.User, error)
	VerifyPassword(user *model.User, password string) bool
	GetUserByAppleUserID(appleUserID string) (*model.User, error)
	CreateAppleUser(appleUserID string) (*model.User, error)
	UpdatePassword(userID string, newPasswordHash string) error
}

// MockUserStore is a mock implementation of UserStoreInterface
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) UpdateUserPreferences(userID string, preferences *model.User) error {
	args := m.Called(userID, preferences)
	return args.Error(0)
}

// --- Implementations for methods needed by AuthService below (for completeness of interface) ---
func (m *MockUserStore) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserStore) CreateEmailUser(email, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserStore) VerifyPassword(user *model.User, password string) bool {
	args := m.Called(user, password)
	return args.Bool(0)
}
func (m *MockUserStore) GetUserByAppleUserID(appleUserID string) (*model.User, error) {
	args := m.Called(appleUserID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserStore) CreateAppleUser(appleUserID string) (*model.User, error) {
	args := m.Called(appleUserID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserStore) UpdatePassword(userID string, newPasswordHash string) error {
	args := m.Called(userID, newPasswordHash)
	return args.Error(0)
}

// Helper to create UserService with mock store
// Assumes UserService.store can be assigned the mock for testing.
func newTestUserService(store UserStoreInterface) *UserService {
	return &UserService{
		store: store, // This assignment requires UserService.store to be of type UserStoreInterface
	}
}

func TestUserService_GetPreferences(t *testing.T) {
	ctx := context.Background()
	userID := "user-pref-123"

	workDuration := 25
	shortBreak := 5
	longBreak := 15
	longInterval := 4

	mockUserWithPrefs := &model.User{
		ID:                 userID,
		Email:              func(s string) *string { return &s }("test@example.com"),
		WorkDuration:       &workDuration,
		ShortBreakDuration: &shortBreak,
		LongBreakDuration:  &longBreak,
		LongBreakInterval:  &longInterval,
		CreatedAt:          time.Now(),
	}

	tests := []struct {
		name                string
		setupMock           func(mockStore *MockUserStore)
		expectedUser        *model.User
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name: "Success",
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPrefs, nil).Once()
			},
			expectedUser: mockUserWithPrefs,
		},
		{
			name: "User Not Found",
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(nil, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrUserNotFound,
		},
		{
			name: "Store GetUserByID Fails",
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(nil, errors.New("db error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			// This assumes NewUserService can take UserStoreInterface or UserService.store is UserStoreInterface
			// userService := NewUserService(mockStore) 
			// For this test structure, we'll use a helper that directly sets the store field:
			userService := &UserService{store: mockStore}


			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			user, err := userService.GetPreferences(ctx, userID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
					}
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdatePreferences(t *testing.T) {
	ctx := context.Background()
	userID := "user-update-pref-123"

	workDuration := 30
	shortBreak := 7
	longBreak := 20
	longInterval := 3
	
	updatedUserMock := &model.User{ // User state after preferences are updated
		ID: userID, 
		WorkDuration: &workDuration,
		ShortBreakDuration: &shortBreak,
		LongBreakDuration: &longBreak,
		LongBreakInterval: &longInterval,
	}


	tests := []struct {
		name                string
		req                 model.UpdateUserPreferencesRequest
		setupMock           func(mockStore *MockUserStore, preferencesMatcher interface{})
		expectedUser        *model.User
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name: "Success - Full Update",
			req: model.UpdateUserPreferencesRequest{
				WorkDuration:       &workDuration,
				ShortBreakDuration: &shortBreak,
				LongBreakDuration:  &longBreak,
				LongBreakInterval:  &longInterval,
			},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				mockStore.On("UpdateUserPreferences", userID, preferencesMatcher).Return(nil).Once()
				mockStore.On("GetUserByID", userID).Return(updatedUserMock, nil).Once()
			},
			expectedUser: updatedUserMock,
		},
		{
			name: "Success - Partial Update (WorkDuration only)",
			req:  model.UpdateUserPreferencesRequest{WorkDuration: &workDuration},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				mockStore.On("UpdateUserPreferences", userID, preferencesMatcher).Return(nil).Once()
				// GetUserByID should return the user with *only* WorkDuration changed from its original state (if any)
				// For simplicity, assume GetUserByID returns the fully formed updatedUserMock,
				// implying the service correctly constructed the partial update for the store.
				partiallyUpdatedUser := &model.User{ID: userID, WorkDuration: &workDuration} // Other fields might be nil or original values
				mockStore.On("GetUserByID", userID).Return(partiallyUpdatedUser, nil).Once() 
			},
			expectedUser: &model.User{ID: userID, WorkDuration: &workDuration},
		},
		{
			name: "Validation Fails - Negative WorkDuration",
			req:  model.UpdateUserPreferencesRequest{WorkDuration: func(i int) *int { return &i }(-10)},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				// No store calls expected
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name: "store.UpdateUserPreferences Fails",
			req:  model.UpdateUserPreferencesRequest{WorkDuration: &workDuration},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				mockStore.On("UpdateUserPreferences", userID, preferencesMatcher).Return(errors.New("db update error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
		{
			name: "store.GetUserByID (after update) Fails",
			req:  model.UpdateUserPreferencesRequest{WorkDuration: &workDuration},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				mockStore.On("UpdateUserPreferences", userID, preferencesMatcher).Return(nil).Once()
				mockStore.On("GetUserByID", userID).Return(nil, errors.New("db get error")).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
		{
			name: "store.GetUserByID (after update) returns nil (User Not Found)",
			req:  model.UpdateUserPreferencesRequest{WorkDuration: &workDuration},
			setupMock: func(mockStore *MockUserStore, preferencesMatcher interface{}) {
				mockStore.On("UpdateUserPreferences", userID, preferencesMatcher).Return(nil).Once()
				mockStore.On("GetUserByID", userID).Return(nil, nil).Once()
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			// userService := NewUserService(mockStore) // Assumes NewUserService takes interface
			userService := &UserService{store: mockStore}


			// Matcher for preferences in UpdateUserPreferences call
			preferencesMatcher := mock.MatchedBy(func(prefs *model.User) bool {
				// Check that only fields present in tt.req are set in prefs, others are nil
				if tt.req.WorkDuration != nil && (prefs.WorkDuration == nil || *prefs.WorkDuration != *tt.req.WorkDuration) { return false }
				if tt.req.WorkDuration == nil && prefs.WorkDuration != nil { return false }
				
				if tt.req.ShortBreakDuration != nil && (prefs.ShortBreakDuration == nil || *prefs.ShortBreakDuration != *tt.req.ShortBreakDuration) { return false }
				if tt.req.ShortBreakDuration == nil && prefs.ShortBreakDuration != nil { return false }

				if tt.req.LongBreakDuration != nil && (prefs.LongBreakDuration == nil || *prefs.LongBreakDuration != *tt.req.LongBreakDuration) { return false }
				if tt.req.LongBreakDuration == nil && prefs.LongBreakDuration != nil { return false }

				if tt.req.LongBreakInterval != nil && (prefs.LongBreakInterval == nil || *prefs.LongBreakInterval != *tt.req.LongBreakInterval) { return false }
				if tt.req.LongBreakInterval == nil && prefs.LongBreakInterval != nil { return false }
				return true
			})


			if tt.setupMock != nil {
				tt.setupMock(mockStore, preferencesMatcher)
			}

			user, err := userService.UpdatePreferences(ctx, userID, tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
					}
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

[end of api/internal/service/user_test.go]
