package store

import (
	"context"
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
		Email:     email,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.db.Exec(`
		INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Email, string(hashedPassword), user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateAppleUser 创建Apple登录用户
func (s *Store) CreateAppleUser(appleUserID string) (*model.User, error) {
	user := &model.User{
		ID:          uuid.New().String(),
		AppleUserID: appleUserID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO users (id, apple_user_id, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.ID, user.AppleUserID, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateOAuthUser 创建OAuth登录用户（Google/Apple等）
func (s *Store) CreateOAuthUser(email, username, provider string) (*model.User, error) {
	user := &model.User{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  username,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO users (id, email, username, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Email, user.Username, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	var passwordHash sql.NullString
	var username sql.NullString
	var appleUserID sql.NullString
	var createdAt, updatedAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, username, password_hash, apple_user_id, is_active, created_at, updated_at
		FROM users
		WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &username, &passwordHash, &appleUserID, &user.IsActive, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if username.Valid {
		user.Username = username.String
	}
	if appleUserID.Valid {
		user.AppleUserID = appleUserID.String
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return &user, nil
}

// GetUserByAppleID 通过Apple ID获取用户
func (s *Store) GetUserByAppleID(appleUserID string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var username sql.NullString
	var createdAt, updatedAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, username, password_hash, apple_user_id, is_active, created_at, updated_at
		FROM users
		WHERE apple_user_id = ?
	`, appleUserID).Scan(&user.ID, &email, &username, new(sql.NullString), &user.AppleUserID, &user.IsActive, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = email.String
	}
	if username.Valid {
		user.Username = username.String
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (s *Store) GetUserByID(id string) (*model.User, error) {
	var user model.User
	var email sql.NullString
	var username sql.NullString
	var appleUserID sql.NullString
	var createdAt, updatedAt time.Time
	err := s.db.QueryRow(`
		SELECT id, email, username, password_hash, apple_user_id, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &email, &username, new(sql.NullString), &appleUserID, &user.IsActive, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = email.String
	}
	if username.Valid {
		user.Username = username.String
	}
	if appleUserID.Valid {
		user.AppleUserID = appleUserID.String
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
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

// GetUserProfile 获取用户资料
func (s *Store) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	query := `SELECT id, user_id, username, avatar, bio, timezone, language, created_at, updated_at 
              FROM user_profiles WHERE user_id = ?`

	var profile model.UserProfile
	var avatar sql.NullString
	var bio sql.NullString

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Username,
		&avatar,
		&bio,
		&profile.Timezone,
		&profile.Language,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &StoreError{Code: "not_found", Message: "profile not found"}
		}
		return nil, err
	}

	if avatar.Valid {
		profile.Avatar = &avatar.String
	}
	if bio.Valid {
		profile.Bio = &bio.String
	}

	return &profile, nil
}

// SaveUserProfile 保存用户资料
func (s *Store) SaveUserProfile(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	// 检查是否存在
	existing, err := s.GetUserProfile(ctx, profile.UserID)
	if err != nil && err.(*StoreError).Code != "not_found" {
		return nil, err
	}

	if existing != nil {
		// 更新
		query := `UPDATE user_profiles SET username = ?, avatar = ?, bio = ?, timezone = ?, 
                  language = ?, updated_at = ? WHERE user_id = ?`
		_, err = s.db.ExecContext(ctx, query, profile.Username, profile.Avatar, profile.Bio,
			profile.Timezone, profile.Language, time.Now(), profile.UserID)
		if err != nil {
			return nil, err
		}
		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
	} else {
		// 创建
		profile.ID = uuid.New().String()
		profile.CreatedAt = time.Now()
		profile.UpdatedAt = time.Now()

		query := `INSERT INTO user_profiles (id, user_id, username, avatar, bio, timezone, language, 
                  created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = s.db.ExecContext(ctx, query, profile.ID, profile.UserID, profile.Username,
			profile.Avatar, profile.Bio, profile.Timezone, profile.Language,
			profile.CreatedAt, profile.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return profile, nil
}

// UpdateUserUsername 更新用户名
func (s *Store) UpdateUserUsername(ctx context.Context, userID, username string) error {
	query := `UPDATE users SET username = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, username, time.Now(), userID)
	return err
}

// UpdateUserPassword 更新用户密码
func (s *Store) UpdateUserPassword(ctx context.Context, userID, newPassword string) error {
	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, string(hashedPassword), time.Now(), userID)
	return err
}

// UpdateUserEmail 更新用户邮箱
func (s *Store) UpdateUserEmail(ctx context.Context, userID, newEmail string) error {
	query := `UPDATE users SET email = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, newEmail, time.Now(), userID)
	return err
}

// UpdateUserAvatar 更新用户头像
func (s *Store) UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error {
	query := `UPDATE users SET avatar = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, avatarURL, time.Now(), userID)
	return err
}

// UpdateUser 更新用户信息
func (s *Store) UpdateUser(user *model.User) error {
	query := `UPDATE users SET email = ?, username = ?, avatar = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, user.Email, user.Username, user.Avatar, time.Now(), user.ID)
	if err != nil {
		return err
	}
	user.UpdatedAt = time.Now()
	return nil
}
