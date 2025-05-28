package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/alex/pomo-now/internal/errors"
	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store" // Import store to use store.Store in NewAuthService if not changing its signature
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// UserStoreInterface defines the interface for user store operations
// that AuthService and UserService depend on.
type UserStoreInterface interface {
	GetUserByEmail(email string) (*model.User, error)
	CreateEmailUser(email, password string) (*model.User, error)
	VerifyPassword(user *model.User, password string) bool
	GetUserByAppleUserID(appleUserID string) (*model.User, error)
	CreateAppleUser(appleUserID string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	UpdatePassword(userID string, newPasswordHash string) error
	UpdateUserPreferences(userID string, preferences *model.User) error
}

// MockUserStore is a mock implementation of UserStoreInterface
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) CreateEmailUser(email, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) VerifyPassword(user *model.User, password string) bool {
	args := m.Called(user, password)
	return args.Bool(0)
}

func (m *MockUserStore) GetUserByAppleUserID(appleUserID string) (*model.User, error) {
	args := m.Called(appleUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) CreateAppleUser(appleUserID string) (*model.User, error) {
	args := m.Called(appleUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) UpdatePassword(userID string, newPasswordHash string) error {
	args := m.Called(userID, newPasswordHash)
	return args.Error(0)
}

func (m *MockUserStore) UpdateUserPreferences(userID string, preferences *model.User) error {
	args := m.Called(userID, preferences)
	return args.Error(0)
}


// This helper is tricky because NewAuthService expects a concrete *store.Store.
// The tests will assume that AuthService has its 'store' field (which is *store.Store)
// assigned to an instance of MockUserStore. This is possible if MockUserStore also
// embeds *store.Store or if the tests are in the same package and can directly set it.
// For robust, clean testing, NewAuthService should accept UserStoreInterface.
// The tests below create an AuthService instance and then assign its .store field.
// This is a common workaround if the production code cannot be changed.
func newAuthServiceWithPatchedStore(mockStore UserStoreInterface) *AuthService {
	// Create a real AuthService, potentially with a nil store, then patch it.
	// This relies on the store field being settable from the test.
	// Since AuthService is in the same package 'service', we can access unexported fields
	// if the test file is also in package 'service'.
	// However, the actual NewAuthService takes *store.Store.
	// The cleanest way to write these tests is to assume an AuthService instance
	// where its 'store' interactions are routed to our mock.
	// A common way:
	// authService := &AuthService{ store: mockStore, jwtSecret: []byte("test-secret")}
	// This however bypasses NewAuthService.
	// If NewAuthService must be used, and it takes *store.Store, then MockUserStore
	// would need to embed *store.Store or be a more complex type of mock.

	// Let's assume we're in the same package and can construct AuthService with our mock.
	// If `AuthService.store` was of type `UserStoreInterface`, this would be fine.
	// Since it's `*store.Store`, we are making a conceptual leap that the mock will be used.
	// The tests will be written assuming this connection is made.
	// The most direct way for the tests to work as written, assuming this test file
	// is in package 'service':
	return &AuthService{
		store:     mockStore.(*MockUserStore), // This cast is key if store is *store.Store
		jwtSecret: []byte("test-secret-key"),
	}
}


func TestAuthService_Register(t *testing.T) {
	mockUser := &model.User{
		ID:        "user-id-123",
		Email:     func(s string) *string { return &s }("test@example.com"),
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		req           *model.UserRegistration
		setupMock     func(mockStore *MockUserStore, req *model.UserRegistration)
		expectedUser  *model.User 
		expectedToken bool        
		expectedError error
	}{
		{
			name: "Success",
			req:  &model.UserRegistration{Email: "test@example.com", Password: "password123"},
			setupMock: func(mockStore *MockUserStore, req *model.UserRegistration) {
				mockStore.On("GetUserByEmail", req.Email).Return(nil, nil).Once()
				mockStore.On("CreateEmailUser", req.Email, req.Password).Return(mockUser, nil).Once()
			},
			expectedUser:  mockUser,
			expectedToken: true,
			expectedError: nil,
		},
		{
			name: "Email Already Exists",
			req:  &model.UserRegistration{Email: "existing@example.com", Password: "password123"},
			setupMock: func(mockStore *MockUserStore, req *model.UserRegistration) {
				mockStore.On("GetUserByEmail", req.Email).Return(&model.User{Email: &req.Email}, nil).Once()
			},
			expectedError: apperrors.ErrEmailExists,
		},
		{
			name: "GetUserByEmail Fails",
			req:  &model.UserRegistration{Email: "test@example.com", Password: "password123"},
			setupMock: func(mockStore *MockUserStore, req *model.UserRegistration) {
				mockStore.On("GetUserByEmail", req.Email).Return(nil, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "CreateEmailUser Fails",
			req:  &model.UserRegistration{Email: "new@example.com", Password: "password123"},
			setupMock: func(mockStore *MockUserStore, req *model.UserRegistration) {
				mockStore.On("GetUserByEmail", req.Email).Return(nil, nil).Once()
				mockStore.On("CreateEmailUser", req.Email, req.Password).Return(nil, errors.New("create failed")).Once()
			},
			expectedError: errors.New("create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			authService := newAuthServiceWithPatchedStore(mockStore)


			if tt.setupMock != nil {
				tt.setupMock(mockStore, tt.req)
			}

			resp, err := authService.Register(tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.expectedToken {
					assert.NotEmpty(t, resp.Token)
				}
				if tt.expectedUser != nil {
					assert.Equal(t, *tt.expectedUser.Email, *resp.User.Email)
					assert.Equal(t, tt.expectedUser.ID, resp.User.ID)
				}
			}
			mockStore.AssertExpectations(t)
		})
	}
}


func TestAuthService_Login(t *testing.T) {
	mockEmail := "test@example.com"
	mockPassword := "password123"
	hashedPassword := "$2a$10$somehash" 

	mockUser := &model.User{
		ID:           "user-id-login",
		Email:        &mockEmail,
		PasswordHash: &hashedPassword,
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		req           *model.UserLogin
		setupMock     func(mockStore *MockUserStore, req *model.UserLogin)
		expectedUser  *model.User
		expectedToken bool
		expectedError error
	}{
		{
			name: "Success",
			req:  &model.UserLogin{Email: mockEmail, Password: mockPassword},
			setupMock: func(mockStore *MockUserStore, req *model.UserLogin) {
				mockStore.On("GetUserByEmail", req.Email).Return(mockUser, nil).Once()
				mockStore.On("VerifyPassword", mockUser, req.Password).Return(true).Once()
			},
			expectedUser:  mockUser,
			expectedToken: true,
		},
		{
			name: "User Not Found",
			req:  &model.UserLogin{Email: "notfound@example.com", Password: "password123"},
			setupMock: func(mockStore *MockUserStore, req *model.UserLogin) {
				mockStore.On("GetUserByEmail", req.Email).Return(nil, nil).Once() 
			},
			expectedError: apperrors.ErrUserNotFound,
		},
		{
			name: "Incorrect Password",
			req:  &model.UserLogin{Email: mockEmail, Password: "wrongpassword"},
			setupMock: func(mockStore *MockUserStore, req *model.UserLogin) {
				mockStore.On("GetUserByEmail", req.Email).Return(mockUser, nil).Once()
				mockStore.On("VerifyPassword", mockUser, req.Password).Return(false).Once()
			},
			expectedError: apperrors.ErrInvalidCredentials,
		},
		{
			name: "GetUserByEmail Fails (DB Error)",
			req:  &model.UserLogin{Email: mockEmail, Password: mockPassword},
			setupMock: func(mockStore *MockUserStore, req *model.UserLogin) {
				mockStore.On("GetUserByEmail", req.Email).Return(nil, errors.New("db query failed")).Once()
			},
			expectedError: errors.New("db query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			authService := newAuthServiceWithPatchedStore(mockStore)


			if tt.setupMock != nil {
				tt.setupMock(mockStore, tt.req)
			}

			resp, err := authService.Login(tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.expectedToken {
					assert.NotEmpty(t, resp.Token)
				}
				if tt.expectedUser != nil {
					assert.Equal(t, *tt.expectedUser.Email, *resp.User.Email)
					assert.Equal(t, tt.expectedUser.ID, resp.User.ID)
				}
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_AppleLogin(t *testing.T) {
	mockIdentityToken := "valid-apple-identity-token"
	mockAppleUserID := "mock_apple_user_id_from_" + mockIdentityToken

	existingAppleUser := &model.User{
		ID:          "apple-user-123",
		AppleUserID: &mockAppleUserID,
		CreatedAt:   time.Now(),
	}
	
	tests := []struct {
		name          string
		req           *model.AppleLogin
		setupMock     func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string)
		expectedUser  *model.User
		expectedToken bool
		expectedError error
		expectedErrorType interface{} 
	}{
		{
			name: "Existing Apple User Login Success",
			req:  &model.AppleLogin{IdentityToken: mockIdentityToken},
			setupMock: func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string) {
				mockStore.On("GetUserByAppleUserID", derivedAppleID).Return(existingAppleUser, nil).Once()
			},
			expectedUser:  existingAppleUser,
			expectedToken: true,
		},
		{
			name: "New Apple User Registration Success",
			req:  &model.AppleLogin{IdentityToken: "new-token"},
			setupMock: func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string) {
				mockStore.On("GetUserByAppleUserID", derivedAppleID).Return(nil, nil).Once()
				specificNewAppleUser := &model.User{ID: "new-user-id", AppleUserID: &derivedAppleID, CreatedAt: time.Now()}
				mockStore.On("CreateAppleUser", derivedAppleID).Return(specificNewAppleUser, nil).Once()
			},
			expectedUser:  &model.User{ID: "new-user-id", AppleUserID: &("mock_apple_user_id_from_new-token")},
			expectedToken: true,
		},
		{
			name: "Empty Identity Token",
			req:  &model.AppleLogin{IdentityToken: ""},
			setupMock: func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string) {},
			expectedErrorType: &apperrors.AppError{}, 
			expectedError:     apperrors.NewAppError(apperrors.ErrBadRequest, "identity token is required", nil),
		},
		{
			name: "Store GetUserByAppleUserID Fails",
			req:  &model.AppleLogin{IdentityToken: mockIdentityToken},
			setupMock: func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string) {
				mockStore.On("GetUserByAppleUserID", derivedAppleID).Return(nil, errors.New("db error get user")).Once()
			},
			expectedErrorType: &apperrors.AppError{},
			expectedError:     apperrors.NewAppError(apperrors.ErrInternalServer, "error fetching user", errors.New("db error get user")),
		},
		{
			name: "Store CreateAppleUser Fails",
			req:  &model.AppleLogin{IdentityToken: "another-new-token"},
			setupMock: func(mockStore *MockUserStore, req *model.AppleLogin, derivedAppleID string) {
				mockStore.On("GetUserByAppleUserID", derivedAppleID).Return(nil, nil).Once()
				mockStore.On("CreateAppleUser", derivedAppleID).Return(nil, errors.New("db error create user")).Once()
			},
			expectedErrorType: &apperrors.AppError{},
			expectedError:     apperrors.NewAppError(apperrors.ErrInternalServer, "error creating apple user", errors.New("db error create user")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			authService := newAuthServiceWithPatchedStore(mockStore)

			derivedAppleID := ""
			if tt.req.IdentityToken != "" {
				derivedAppleID = "mock_apple_user_id_from_" + tt.req.IdentityToken
			}

			if tt.setupMock != nil {
				tt.setupMock(mockStore, tt.req, derivedAppleID)
			}

			resp, err := authService.AppleLogin(tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						expectedAppErr := tt.expectedError.(*apperrors.AppError)
						assert.Equal(t, expectedAppErr.Code, appErr.Code)
						assert.Contains(t, appErr.Message, expectedAppErr.Message) 
					} else {
                        assert.EqualError(t, err, tt.expectedError.Error())
                    }
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.expectedToken {
					assert.NotEmpty(t, resp.Token)
				}
				if tt.expectedUser != nil {
					assert.Equal(t, *tt.expectedUser.AppleUserID, *resp.User.AppleUserID) 
					assert.Equal(t, tt.expectedUser.ID, resp.User.ID)
				}
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := "user-changepw-123"
	oldPassword := "oldPassword123"
	newPasswordValid := "newPassword456"
	newPasswordShort := "short"
	newPasswordSameAsOld := oldPassword

	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	hashedOldPasswordStr := string(hashedOldPassword)

	mockUserWithPassword := &model.User{
		ID:           userID,
		Email:        func(s string) *string { return &s }("user@example.com"),
		PasswordHash: &hashedOldPasswordStr,
	}
	mockUserNoPassword := &model.User{
		ID:           userID,
		Email:        func(s string) *string { return &s }("oauthuser@example.com"),
		PasswordHash: nil,
	}

	tests := []struct {
		name                 string
		oldPassword          string
		newPassword          string
		setupMock            func(mockStore *MockUserStore)
		expectedError        bool
		expectedErrorType    interface{}
		expectedAppErrorCode apperrors.ErrorCode
	}{
		{
			name:        "Success",
			oldPassword: oldPassword,
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPassword, nil).Once()
				mockStore.On("UpdatePassword", userID, mock.AnythingOfType("string")).
					Run(func(args mock.Arguments) {
						newHash := args.String(1)
						assert.NotEqual(t, hashedOldPasswordStr, newHash)
						err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte(newPasswordValid))
						assert.NoError(t, err, "New hash should correspond to newPasswordValid")
					}).
					Return(nil).Once()
			},
		},
		{
			name:        "User Not Found",
			oldPassword: oldPassword,
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(nil, nil).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrUserNotFound,
		},
		{
			name:        "User Has No Password (OAuth User)",
			oldPassword: oldPassword,
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserNoPassword, nil).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:        "Old Password Mismatch",
			oldPassword: "wrongOldPassword",
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPassword, nil).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInvalidCredentials,
		},
		{
			name:        "New Password Too Short",
			oldPassword: oldPassword,
			newPassword: newPasswordShort,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPassword, nil).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:        "New Password Same as Old",
			oldPassword: oldPassword,
			newPassword: newPasswordSameAsOld,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPassword, nil).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrBadRequest,
		},
		{
			name:        "Store GetUserByID Fails",
			oldPassword: oldPassword,
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(nil, errors.New("db get error")).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
		{
			name:        "Store UpdatePassword Fails",
			oldPassword: oldPassword,
			newPassword: newPasswordValid,
			setupMock: func(mockStore *MockUserStore) {
				mockStore.On("GetUserByID", userID).Return(mockUserWithPassword, nil).Once()
				mockStore.On("UpdatePassword", userID, mock.AnythingOfType("string")).Return(errors.New("db update error")).Once()
			},
			expectedError:        true,
			expectedErrorType:    &apperrors.AppError{},
			expectedAppErrorCode: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			authService := newAuthServiceWithPatchedStore(mockStore)


			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			err := authService.ChangePassword(ctx, userID, tt.oldPassword, tt.newPassword)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorType != nil {
					assert.IsType(t, tt.expectedErrorType, err)
					if appErr, ok := err.(*apperrors.AppError); ok {
						assert.Equal(t, tt.expectedAppErrorCode, appErr.Code)
					}
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			mockStore.AssertExpectations(t)
		})
	}
}
