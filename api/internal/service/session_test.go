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

// SessionStoreInterface defines the interface for session store operations
// that SessionService depends on.
type SessionStoreInterface interface {
	CreateSession(session *model.Session) error
	GetPomodoroStatsByUserID(userID string, startTime time.Time, endTime time.Time) (totalFocusSessions int, totalFocusMinutes int, err error)
	GetCompletedTasksCountByUserID(userID string, startTime time.Time, endTime time.Time) (int, error)
	// Add other methods if SessionService depends on them directly for these functionalities.
	// For example, ListSessions, GetSession, etc., if they were part of the tested methods.
	// For RecordSession and GetStatistics, the above are sufficient.
}

// MockSessionStore is a mock implementation of SessionStoreInterface
type MockSessionStore struct {
	mock.Mock
}

func (m *MockSessionStore) CreateSession(session *model.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockSessionStore) GetPomodoroStatsByUserID(userID string, startTime time.Time, endTime time.Time) (int, int, error) {
	args := m.Called(userID, startTime, endTime)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockSessionStore) GetCompletedTasksCountByUserID(userID string, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

// Helper to create SessionService with mock store
// Assumes SessionService.store can be assigned the mock for testing.
func newTestSessionService(store SessionStoreInterface) *SessionService {
	return &SessionService{
		store: store,
	}
}

func TestSessionService_RecordSession(t *testing.T) {
	ctx := context.Background()
	userID := "user-record-session"
	taskID := "task-for-session"

	tests := []struct {
		name                string
		req                 model.RecordSessionRequest
		setupMock           func(mockStore *MockSessionStore, sessionMatcher interface{})
		checkEndedAtNearNow bool // For cases where EndedAt is defaulted
		expectedError       bool
		expectedErrorType   interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name: "Success (EndedAt Provided)",
			req: model.RecordSessionRequest{
				TaskID:          &taskID,
				DurationMinutes: 25,
				Type:            model.SessionTypeFocus,
				EndedAt:         func() *time.Time { t := time.Now().Add(-5 * time.Minute); return &t }(),
			},
			setupMock: func(mockStore *MockSessionStore, sessionMatcher interface{}) {
				mockStore.On("CreateSession", sessionMatcher).Return(nil).Once()
			},
		},
		{
			name: "Success (EndedAt Not Provided)",
			req: model.RecordSessionRequest{
				TaskID:          nil,
				DurationMinutes: 5,
				Type:            model.SessionTypeShortBreak,
			},
			setupMock: func(mockStore *MockSessionStore, sessionMatcher interface{}) {
				mockStore.On("CreateSession", sessionMatcher).Return(nil).Once()
			},
			checkEndedAtNearNow: true,
		},
		{
			name: "Invalid DurationMinutes - Zero",
			req: model.RecordSessionRequest{
				DurationMinutes: 0,
				Type:            model.SessionTypeFocus,
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name: "Invalid DurationMinutes - Negative",
			req: model.RecordSessionRequest{
				DurationMinutes: -5,
				Type:            model.SessionTypeFocus,
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name: "Invalid Type",
			req: model.RecordSessionRequest{
				DurationMinutes: 25,
				Type:            "invalid_type",
			},
			expectedError:       true,
			expectedErrorType:   &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name: "Store CreateSession Fails",
			req: model.RecordSessionRequest{
				DurationMinutes: 15,
				Type:            model.SessionTypeLongBreak,
			},
			setupMock: func(mockStore *MockSessionStore, sessionMatcher interface{}) {
				mockStore.On("CreateSession", sessionMatcher).Return(errors.New("db insert error")).Once()
			},
			expectedError: true,
			// Service currently returns raw db error. Ideally, wrap in AppError.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockSessionStore)
			sessionService := newTestSessionService(mockStore)

			var expectedEndedAt time.Time
			if tt.req.EndedAt != nil {
				expectedEndedAt = tt.req.EndedAt.UTC()
			}

			sessionMatcher := mock.MatchedBy(func(session *model.Session) bool {
				if session.UserID != userID ||
					session.DurationMinutes != tt.req.DurationMinutes ||
					session.Type != tt.req.Type {
					return false
				}
				if tt.req.TaskID == nil && session.TaskID != nil { return false }
				if tt.req.TaskID != nil && (session.TaskID == nil || *session.TaskID != *tt.req.TaskID) { return false }
				
				if tt.checkEndedAtNearNow {
					return !session.EndedAt.IsZero() && time.Since(session.EndedAt.UTC()) < 2*time.Second
				}
				if tt.req.EndedAt != nil {
					return session.EndedAt.Equal(expectedEndedAt)
				}
				return true // Should only reach here if EndedAt is not provided and not checked for "near now"
			})
			
			if tt.setupMock != nil {
				tt.setupMock(mockStore, sessionMatcher)
			}

			createdSession, err := sessionService.RecordSession(ctx, userID, tt.req)

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
				assert.NotNil(t, createdSession)
				assert.Equal(t, userID, createdSession.UserID)
				assert.Equal(t, tt.req.DurationMinutes, createdSession.DurationMinutes)
				assert.Equal(t, tt.req.Type, createdSession.Type)
				if tt.req.TaskID != nil {
					assert.NotNil(t, createdSession.TaskID)
					assert.Equal(t, *tt.req.TaskID, *createdSession.TaskID)
				} else {
					assert.Nil(t, createdSession.TaskID)
				}
				assert.NotEmpty(t, createdSession.ID)

				if tt.checkEndedAtNearNow {
					assert.WithinDuration(t, time.Now().UTC(), createdSession.EndedAt.UTC(), 2*time.Second)
				} else if tt.req.EndedAt != nil {
					assert.True(t, expectedEndedAt.Equal(createdSession.EndedAt.UTC()))
				}
			}
			mockStore.AssertExpectations(t)
		})
	}
}


func TestSessionService_GetStatistics(t *testing.T) {
	ctx := context.Background()
	userID := "user-stats-123"

	// Helper for calculating expected start/end dates for test assertions
	calculatePeriodRange := func(period string, now time.Time) (time.Time, time.Time) {
		var startTime, endTime time.Time
		switch period {
		case "today":
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			endTime = startTime.AddDate(0, 0, 1).Add(-time.Nanosecond)
		case "week":
			weekday := now.Weekday()
			daysToSubtract := int(weekday) - int(time.Monday)
			if daysToSubtract < 0 { daysToSubtract += 7 }
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysToSubtract)
			endTime = startTime.AddDate(0, 0, 7).Add(-time.Nanosecond)
		case "month":
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			endTime = startTime.AddDate(0, 1, 0).Add(-time.Nanosecond)
		default: // covers "invalid_period_defaults_to_today"
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			endTime = startTime.AddDate(0, 0, 1).Add(-time.Nanosecond)
		}
		return startTime, endTime
	}
	
	// Use a fixed "now" for predictable period calculations in tests
	// Note: If the service's `time.Now()` is not controlled, tests for period ranges might be flaky.
	// For this test, we'll assume the logic in GetStatistics for date calculation is correct and deterministic
	// and we'll assert against ranges calculated by our helper here which mimics that logic.
	fixedNow := time.Now().UTC() 


	tests := []struct {
		name                  string
		period                string
		mockFocusSessions     int
		mockFocusMinutes      int
		mockTasksCompleted    int
		mockPomodoroStatsErr  error
		mockTasksCompletedErr error
		expectedError         bool
		expectedErrorType     interface{}
		expectedAppErrorCode   apperrors.ErrorCode
	}{
		{
			name:               "Success - today",
			period:             "today",
			mockFocusSessions:  5,
			mockFocusMinutes:   125,
			mockTasksCompleted: 2,
		},
		{
			name:               "Success - week",
			period:             "week",
			mockFocusSessions:  20,
			mockFocusMinutes:   500,
			mockTasksCompleted: 10,
		},
		{
			name:               "Success - month",
			period:             "month",
			mockFocusSessions:  80,
			mockFocusMinutes:   2000,
			mockTasksCompleted: 30,
		},
		{
			name:               "Success - invalid period (defaults to today)",
			period:             "invalid_period", // Service logic defaults this to "today"
			mockFocusSessions:  3,
			mockFocusMinutes:   75,
			mockTasksCompleted: 1,
		},
		{
			name:                 "Error from GetPomodoroStatsByUserID",
			period:               "today",
			mockPomodoroStatsErr: errors.New("db pomodoro error"),
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode:  apperrors.ErrInternalServer,
		},
		{
			name:                  "Error from GetCompletedTasksCountByUserID",
			period:                "today",
			mockFocusSessions:     2,
			mockFocusMinutes:      50,
			mockTasksCompletedErr: errors.New("db tasks error"),
			expectedError:         true,
			expectedErrorType:     &apperrors.AppError{},
			expectedAppErrorCode:   apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockSessionStore)
			sessionService := newTestSessionService(mockStore)

			expectedPeriod := tt.period
			if tt.period == "invalid_period" { // Service defaults invalid period to "today"
				expectedPeriod = "today"
			}
			expectedStartTime, expectedEndTime := calculatePeriodRange(expectedPeriod, fixedNow)


			// Setup mocks
			if tt.mockPomodoroStatsErr == nil {
				// Use mock.Anything for time.Time if exact match is tricky due to nanoseconds or minor variations
				// However, since our service calculates these deterministically, we can try to match them.
				mockStore.On("GetPomodoroStatsByUserID", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(tt.mockFocusSessions, tt.mockFocusMinutes, nil).
					// Ensure the time arguments are within a small delta of expected, or match exactly.
					// This is a more robust way to check time arguments:
					// .MatchedBy(func(uid string, st time.Time, et time.Time) bool {
					// 	return uid == userID && st.Unix() == expectedStartTime.Unix() && et.Unix() == expectedEndTime.Unix()
					// }).
					Once() // Expect once if no error before it
			} else {
				mockStore.On("GetPomodoroStatsByUserID", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
					Return(0, 0, tt.mockPomodoroStatsErr).Once()
			}

			if tt.mockPomodoroStatsErr == nil { // Only setup next mock if first one is expected to succeed
				if tt.mockTasksCompletedErr == nil {
					mockStore.On("GetCompletedTasksCountByUserID", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
						Return(tt.mockTasksCompleted, nil).Once()
				} else {
					mockStore.On("GetCompletedTasksCountByUserID", userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
						Return(0, tt.mockTasksCompletedErr).Once()
				}
			}


			stats, err := sessionService.GetStatistics(ctx, userID, tt.period)

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
				assert.NotNil(t, stats)
				assert.Equal(t, tt.mockFocusSessions, stats.TotalFocusSessions)
				assert.Equal(t, tt.mockFocusMinutes, stats.TotalFocusMinutes)
				assert.Equal(t, tt.mockTasksCompleted, stats.TotalTasksCompleted)
				assert.Equal(t, expectedPeriod, stats.Period)
				
				// Check start and end dates. Allow for slight variations if time.Now() in service is not mocked.
				// For deterministic test, service's time.Now() should be injectable or mocked.
				// Here, we compare against our helper's calculation using a fixedNow.
				// The service GetStatistics also uses time.Now().UTC() internally.
				// If the test runs very close to a day/week/month boundary, it might be flaky.
				// For robustness, one might inject a clock into the service.
				// For this test, we'll assume the service's date calculations are correct and compare.
				// The key is that the store is called with *some* date range, and the service uses that range.
				assert.WithinDuration(t, expectedStartTime, stats.StartDate, time.Second, "StartDate mismatch")
				assert.WithinDuration(t, expectedEndTime, stats.EndDate, time.Second, "EndDate mismatch")

			}
			mockStore.AssertExpectations(t)
		})
	}
}

[end of api/internal/service/session_test.go]
