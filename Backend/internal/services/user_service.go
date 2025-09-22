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

// StorageStatistics represents comprehensive storage statistics for a user
type StorageStatistics struct {
	TotalStorage    int64   `json:"total_storage"`    // Deduplicated storage used in bytes
	OriginalStorage int64   `json:"original_storage"` // Storage without deduplication in bytes
	StorageQuota    int64   `json:"storage_quota"`    // User's storage quota in bytes
	FileCount       int     `json:"file_count"`       // Total number of files owned
	DuplicateCount  int     `json:"duplicate_count"`  // Number of duplicate files avoided
	Savings         Savings `json:"savings"`          // Savings from deduplication
}

type Savings struct {
	Bytes      int64   `json:"bytes"`      // Bytes saved through deduplication
	Percentage float64 `json:"percentage"` // Percentage saved (0-100)
}

// GetStorageStatistics calculates comprehensive storage statistics for a user
func (s *UserService) GetStorageStatistics(userID string) (*StorageStatistics, error) {
	var stats StorageStatistics
	var err error

	// Get user's quota
	var user models.User
	err = s.db.Select("storage_quota").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	stats.StorageQuota = user.StorageQuota

	// Get total file count for this user
	var fileCount int64
	err = s.db.Model(&models.UserFile{}).Where("user_id = ?", userID).Count(&fileCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count user files: %w", err)
	}
	stats.FileCount = int(fileCount)

	// Calculate deduplicated storage (what the user is actually using)
	// This is the sum of unique file sizes that this user has
	type DeduplicatedResult struct {
		TotalSize int64
	}
	var deduplicatedResult DeduplicatedResult

	err = s.db.Model(&models.FileHash{}).
		Select("COALESCE(SUM(DISTINCT file_hashes.size), 0) as total_size").
		Joins("JOIN user_files ON file_hashes.hash = user_files.file_hash").
		Where("user_files.user_id = ?", userID).
		Scan(&deduplicatedResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate deduplicated storage: %w", err)
	}
	stats.TotalStorage = deduplicatedResult.TotalSize

	// Calculate original storage (sum of all file sizes without deduplication)
	// This is what the storage would be if we didn't deduplicate
	type OriginalResult struct {
		TotalSize int64
	}
	var originalResult OriginalResult

	err = s.db.Model(&models.UserFile{}).
		Select("COALESCE(SUM(file_hashes.size), 0) as total_size").
		Joins("JOIN file_hashes ON user_files.file_hash = file_hashes.hash").
		Where("user_files.user_id = ?", userID).
		Scan(&originalResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate original storage: %w", err)
	}
	stats.OriginalStorage = originalResult.TotalSize

	// Calculate duplicate count (files that would have been duplicates)
	// This is the number of user files minus the number of unique hashes
	type UniqueHashCount struct {
		Count int64
	}
	var uniqueHashCount UniqueHashCount

	err = s.db.Model(&models.UserFile{}).
		Select("COUNT(DISTINCT file_hash) as count").
		Where("user_id = ?", userID).
		Scan(&uniqueHashCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique hashes: %w", err)
	}

	stats.DuplicateCount = stats.FileCount - int(uniqueHashCount.Count)
	if stats.DuplicateCount < 0 {
		stats.DuplicateCount = 0 // Safety check
	}

	// Calculate savings
	stats.Savings.Bytes = stats.OriginalStorage - stats.TotalStorage
	if stats.Savings.Bytes < 0 {
		stats.Savings.Bytes = 0 // Safety check
	}

	if stats.OriginalStorage > 0 {
		stats.Savings.Percentage = float64(stats.Savings.Bytes) / float64(stats.OriginalStorage) * 100
	} else {
		stats.Savings.Percentage = 0
	}

	// Ensure percentage is within valid range
	if stats.Savings.Percentage < 0 {
		stats.Savings.Percentage = 0
	} else if stats.Savings.Percentage > 100 {
		stats.Savings.Percentage = 100
	}

	return &stats, nil
}
