package services

import (
	"fmt"
	"time"

	"filevault-backend/internal/config"
	"filevault-backend/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{
		db:  db,
		cfg: cfg,
	}
}

// GetOrCreateUser finds existing user or creates new one based on Clerk user ID
func (s *UserService) GetOrCreateUser(clerkUserID, email, firstName, lastName string) (*models.User, error) {
	var user models.User

	err := s.db.Where("id = ?", clerkUserID).First(&user).Error
	if err == nil {
		// User exists, update role in case it changed
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Create new user with configurable storage quota
	user = models.User{
		ID:           clerkUserID,
		Role:         models.UserRoleUser,
		StorageQuota: s.cfg.DefaultStorageQuotaMB * 1024 * 1024, // Convert MB to bytes
		StorageUsed:  0,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUser retrieves user by ID
func (s *UserService) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// UpdateUserRole updates user role (admin function)
func (s *UserService) UpdateUserRole(userID string, role models.UserRole) error {
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	return nil
}

// UpdateStorageUsed updates user's storage usage
func (s *UserService) UpdateStorageUsed(userID string, sizeDelta int64) error {
	err := s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("storage_used", gorm.Expr("storage_used + ?", sizeDelta)).Error
	if err != nil {
		return fmt.Errorf("failed to update storage used: %w", err)
	}
	return nil
}

// CheckStorageQuota checks if user has enough quota for additional storage
func (s *UserService) CheckStorageQuota(userID string, additionalSize int64) error {
	var user models.User
	err := s.db.Select("storage_quota", "storage_used").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return fmt.Errorf("failed to get user storage info: %w", err)
	}

	if user.StorageUsed+additionalSize > user.StorageQuota {
		return fmt.Errorf("storage quota exceeded: have %d bytes, need %d bytes, quota is %d bytes",
			user.StorageUsed, additionalSize, user.StorageQuota)
	}

	return nil
}

// GetUserStorageInfo returns user's storage usage and quota
func (s *UserService) GetUserStorageInfo(userID string) (used, quota int64, err error) {
	var user models.User
	err = s.db.Select("storage_quota", "storage_used").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user storage info: %w", err)
	}
	return user.StorageUsed, user.StorageQuota, nil
}

// ListUsers returns paginated list of users (admin function)
func (s *UserService) ListUsers(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Get total count
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated users
	err := s.db.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateStorageQuota updates user's storage quota (admin function)
func (s *UserService) UpdateStorageQuota(userID string, quotaMB int64) error {
	// Validate quota doesn't exceed maximum
	if quotaMB > s.cfg.MaxStorageQuotaMB {
		return fmt.Errorf("storage quota cannot exceed %d MB", s.cfg.MaxStorageQuotaMB)
	}

	quotaBytes := quotaMB * 1024 * 1024
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("storage_quota", quotaBytes).Error
	if err != nil {
		return fmt.Errorf("failed to update storage quota: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user (admin function)
func (s *UserService) DeleteUser(userID string) error {
	err := s.db.Where("id = ?", userID).Delete(&models.User{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
