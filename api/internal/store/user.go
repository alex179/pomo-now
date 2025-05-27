package store

import (
	"database/sql"
	"time"

	"github.com/alex/pomo-now/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser 创建新用户
func (s *Store) CreateUser(email, password string) (*model.User, error) {
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:        uuid.New().String(),
		Email:     &email,
		CreatedAt: time.Now(),
	}

	_, err = s.db.Exec(`
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, user.ID, user.Email, string(hashedPassword), user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateAppleUser 创建Apple登录用户
func (s *Store) CreateAppleUser(appleUserID string) (*model.User, error) {
	user := &model.User{
		ID:          uuid.New().String(),
		AppleUserID: &appleUserID,
		CreatedAt:   time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO users (id, apple_user_id, created_at)
		VALUES (?, ?, ?)
	`, user.ID, user.AppleUserID, user.CreatedAt)
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

// GetUserByAppleID 通过Apple ID获取用户
func (s *Store) GetUserByAppleID(appleUserID string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var createdAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at
		FROM users
		WHERE apple_user_id = ?
	`, appleUserID).Scan(&user.ID, &email, new(sql.NullString), &user.AppleUserID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = &email.String
	}
	user.CreatedAt = createdAt
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (s *Store) GetUserByID(id string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var appleUserID sql.NullString
	var createdAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, apple_user_id, created_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &email, new(sql.NullString), &appleUserID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = &email.String
	}
	if appleUserID.Valid {
		user.AppleUserID = &appleUserID.String
	}
	user.CreatedAt = createdAt
	return &user, nil
}

// VerifyPassword 验证用户密码
func (s *Store) VerifyPassword(user *model.User, password string) bool {
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT password_hash FROM users WHERE id = ?
	`, user.ID).Scan(&passwordHash)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
