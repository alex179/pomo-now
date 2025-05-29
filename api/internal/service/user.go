package service

import (
	"context"
	"errors"

	"github.com/alex/pomo-now/internal/model"
	"github.com/alex/pomo-now/internal/store"
)

// UserService 用户服务
type UserService struct {
	store *store.Store
}

// NewUserService 创建用户服务
func NewUserService(store *store.Store) *UserService {
	return &UserService{store: store}
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID string) (*model.UserInfoResponse, error) {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 获取用户资料
	profile, err := s.store.GetUserProfile(ctx, userID)
	if err != nil {
		// 如果没有资料，profile为nil
		profile = nil
	}

	return &model.UserInfoResponse{
		ID:        user.ID,
		Email:     &user.Email,
		Username:  &user.Username,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		Profile:   profile,
	}, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.UserInfoResponse, error) {
	// 更新用户基本信息
	if req.Username != nil {
		err := s.store.UpdateUserUsername(ctx, userID, *req.Username)
		if err != nil {
			return nil, err
		}
	}

	// 更新或创建用户资料
	profile, err := s.store.GetUserProfile(ctx, userID)
	if err != nil {
		// 创建新资料
		profile = &model.UserProfile{
			UserID:   userID,
			Username: "",
			Timezone: "Asia/Shanghai",
			Language: "zh-CN",
		}
	}

	if req.Username != nil {
		profile.Username = *req.Username
	}
	if req.Bio != nil {
		profile.Bio = req.Bio
	}
	if req.Timezone != nil {
		profile.Timezone = *req.Timezone
	}
	if req.Language != nil {
		profile.Language = *req.Language
	}

	profile, err = s.store.SaveUserProfile(ctx, profile)
	if err != nil {
		return nil, err
	}

	// 返回更新后的用户信息
	return s.GetUserInfo(ctx, userID)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID string, currentPassword, newPassword string) error {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 验证当前密码
	if !s.store.VerifyPassword(user, currentPassword) {
		return errors.New("invalid current password")
	}

	// 更新密码
	return s.store.UpdateUserPassword(ctx, userID, newPassword)
}

// UpdateEmail 修改邮箱
func (s *UserService) UpdateEmail(ctx context.Context, userID string, newEmail, password string) error {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 验证密码
	if !s.store.VerifyPassword(user, password) {
		return errors.New("invalid password")
	}

	// 检查新邮箱是否已被使用
	existingUser, err := s.store.GetUserByEmail(newEmail)
	if err != nil {
		return err
	}
	if existingUser != nil && existingUser.ID != userID {
		return errors.New("email already exists")
	}

	// 更新邮箱
	return s.store.UpdateUserEmail(ctx, userID, newEmail)
}
