package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser inserts a new user record into the database.
// It expects the User model to be populated, including a hashed password if applicable.
func (s *Store) CreateUser(user *model.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT INTO users (id, email, password_hash, apple_user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.ID, user.Email, user.PasswordHash, user.AppleUserID, user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// CreateEmailUser creates a new user with email and password.
// This is a convenience function that hashes the password.
func (s *Store) CreateEmailUser(email, password string) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Email:        &email,
		PasswordHash: func(s string) *string { return &s }(string(hashedPassword)),
		CreatedAt:    time.Now(),
	}

	err = s.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateAppleUser creates a new Apple sign-in user.
func (s *Store) CreateAppleUser(appleUserID string) (*model.User, error) {
	user := &model.User{
		ID:          uuid.New().String(),
		AppleUserID: &appleUserID,
		CreatedAt:   time.Now(),
	}

	err := s.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	var passwordHash, appleUserID sql.NullString
	var workDuration, shortBreakDuration, longBreakDuration, longBreakInterval sql.NullInt32 // Use NullInt32 for nullable INTs

	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at,
		       work_duration, short_break_duration, long_break_duration, long_break_interval
		FROM users
		WHERE email = ?
	`, email).Scan(
		&user.ID, &user.Email, &passwordHash, &appleUserID, &user.CreatedAt,
		&workDuration, &shortBreakDuration, &longBreakDuration, &longBreakInterval,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if passwordHash.Valid { user.PasswordHash = &passwordHash.String }
	if appleUserID.Valid { user.AppleUserID = &appleUserID.String }
	if workDuration.Valid { val := int(workDuration.Int32); user.WorkDuration = &val }
	if shortBreakDuration.Valid { val := int(shortBreakDuration.Int32); user.ShortBreakDuration = &val }
	if longBreakDuration.Valid { val := int(longBreakDuration.Int32); user.LongBreakDuration = &val }
	if longBreakInterval.Valid { val := int(longBreakInterval.Int32); user.LongBreakInterval = &val }
	
	return &user, nil
}

// GetUserByAppleUserID 通过Apple ID获取用户
func (s *Store) GetUserByAppleUserID(appleUserID string) (*model.User, error) {
	var user model.User
	var email, passwordHash sql.NullString
	var workDuration, shortBreakDuration, longBreakDuration, longBreakInterval sql.NullInt32

	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at,
		       work_duration, short_break_duration, long_break_duration, long_break_interval
		FROM users
		WHERE apple_user_id = ?
	`, appleUserID).Scan(
		&user.ID, &email, &passwordHash, &user.AppleUserID, &user.CreatedAt,
		&workDuration, &shortBreakDuration, &longBreakDuration, &longBreakInterval,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if email.Valid { user.Email = &email.String }
	if passwordHash.Valid { user.PasswordHash = &passwordHash.String }
	if workDuration.Valid { val := int(workDuration.Int32); user.WorkDuration = &val }
	if shortBreakDuration.Valid { val := int(shortBreakDuration.Int32); user.ShortBreakDuration = &val }
	if longBreakDuration.Valid { val := int(longBreakDuration.Int32); user.LongBreakDuration = &val }
	if longBreakInterval.Valid { val := int(longBreakInterval.Int32); user.LongBreakInterval = &val }

	return &user, nil
}

// GetUserByID 通过ID获取用户
func (s *Store) GetUserByID(id string) (*model.User, error) {
	var user model.User
	var email, passwordHash, appleUserID sql.NullString
	var workDuration, shortBreakDuration, longBreakDuration, longBreakInterval sql.NullInt32
	
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at,
		       work_duration, short_break_duration, long_break_duration, long_break_interval
		FROM users
		WHERE id = ?
	`, id).Scan(
		&user.ID, &email, &passwordHash, &appleUserID, &user.CreatedAt,
		&workDuration, &shortBreakDuration, &longBreakDuration, &longBreakInterval,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if email.Valid { user.Email = &email.String }
	if passwordHash.Valid { user.PasswordHash = &passwordHash.String }
	if appleUserID.Valid { user.AppleUserID = &appleUserID.String }
	if workDuration.Valid { val := int(workDuration.Int32); user.WorkDuration = &val }
	if shortBreakDuration.Valid { val := int(shortBreakDuration.Int32); user.ShortBreakDuration = &val }
	if longBreakDuration.Valid { val := int(longBreakDuration.Int32); user.LongBreakDuration = &val }
	if longBreakInterval.Valid { val := int(longBreakInterval.Int32); user.LongBreakInterval = &val }
	
	return &user, nil
}

// UpdateUserPreferences updates specific preference fields for a user.
func (s *Store) UpdateUserPreferences(userID string, preferences *model.User) error {
	// Build the SET part of the query dynamically
	// This is a simplified example. For many fields, a more robust query builder might be better.
	var querySetParts []string
	var args []interface{}

	if preferences.WorkDuration != nil {
		querySetParts = append(querySetParts, "work_duration = ?")
		args = append(args, *preferences.WorkDuration)
	}
	if preferences.ShortBreakDuration != nil {
		querySetParts = append(querySetParts, "short_break_duration = ?")
		args = append(args, *preferences.ShortBreakDuration)
	}
	if preferences.LongBreakDuration != nil {
		querySetParts = append(querySetParts, "long_break_duration = ?")
		args = append(args, *preferences.LongBreakDuration)
	}
	if preferences.LongBreakInterval != nil {
		querySetParts = append(querySetParts, "long_break_interval = ?")
		args = append(args, *preferences.LongBreakInterval)
	}

	if len(querySetParts) == 0 {
		return nil // No preferences to update
	}

	query := "UPDATE users SET " + string(querySetParts[0])
	for i := 1; i < len(querySetParts); i++ {
		query += ", " + string(querySetParts[i])
	}
	query += " WHERE id = ?"
	args = append(args, userID)

	_, err := s.db.Exec(query, args...)
	return err
}

// UpdatePassword updates the password_hash for a given user.
func (s *Store) UpdatePassword(userID string, newPasswordHash string) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET password_hash = ?
		WHERE id = ?
	`, newPasswordHash, userID)
	return err
}


// VerifyPassword 验证用户密码
func (s *Store) VerifyPassword(user *model.User, password string) bool {
	// Ensure user.PasswordHash is populated, if not, try to fetch it.
	// However, it's better if the user object passed to VerifyPassword already has PasswordHash.
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		var dbPasswordHash string
		err := s.db.QueryRow(`
			SELECT password_hash FROM users WHERE id = ? AND password_hash IS NOT NULL
		`, user.ID).Scan(&dbPasswordHash)
		if err != nil { // Handles sql.ErrNoRows or other errors
			return false
		}
		return bcrypt.CompareHashAndPassword([]byte(dbPasswordHash), []byte(password)) == nil
	}

	return bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) == nil
}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	var passwordHash sql.NullString
	var appleUserID sql.NullString
	var createdAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at
		FROM users
		WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &passwordHash, &appleUserID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if appleUserID.Valid {
		user.AppleUserID = &appleUserID.String
	}
	user.CreatedAt = createdAt
	return &user, nil
}

// GetUserByAppleUserID 通过Apple ID获取用户 (renamed from GetUserByAppleID)
func (s *Store) GetUserByAppleUserID(appleUserID string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var passwordHash sql.NullString // To hold PasswordHash from DB
	var createdAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at
		FROM users
		WHERE apple_user_id = ?
	`, appleUserID).Scan(&user.ID, &email, &passwordHash, &user.AppleUserID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = &email.String
	}
	if passwordHash.Valid { // Check if PasswordHash is not null
		user.PasswordHash = &passwordHash.String
	}
	user.CreatedAt = createdAt
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (s *Store) GetUserByID(id string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var passwordHash sql.NullString // To hold PasswordHash from DB
	var appleUserID sql.NullString
	var createdAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &email, &passwordHash, &appleUserID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = &email.String
	}
	if passwordHash.Valid { // Check if PasswordHash is not null
		user.PasswordHash = &passwordHash.String
	}
	if appleUserID.Valid {
		user.AppleUserID = &appleUserID.String
	}
	user.CreatedAt = createdAt
	return &user, nil
}

// VerifyPassword 验证用户密码
func (s *Store) VerifyPassword(user *model.User, password string) bool {
	// Ensure user.PasswordHash is populated, if not, try to fetch it.
	// However, it's better if the user object passed to VerifyPassword already has PasswordHash.
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		var dbPasswordHash string
		err := s.db.QueryRow(`
			SELECT password_hash FROM users WHERE id = ? AND password_hash IS NOT NULL
		`, user.ID).Scan(&dbPasswordHash)
		if err != nil { // Handles sql.ErrNoRows or other errors
			return false
		}
		return bcrypt.CompareHashAndPassword([]byte(dbPasswordHash), []byte(password)) == nil
	}

	return bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) == nil
}
