package services

import (
	"context"
	"fmt"
	"time"

	"filevault-backend/internal/models"
	"filevault-backend/internal/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileService struct {
	db      *gorm.DB
	storage *storage.MinIOStorage
}

func NewFileService(db *gorm.DB, storage *storage.MinIOStorage) *FileService {
	return &FileService{
		db:      db,
		storage: storage,
	}
}

// GeneratePresignedUploadURL generates a presigned URL for file upload
func (s *FileService) GeneratePresignedUploadURL(userID, filename, fileHash string, size int64, mimeType string) (*PresignedUploadResponse, error) {
	// Use the hash provided by frontend for the object key
	tempKey := fmt.Sprintf("temp/%s/%s", userID, fileHash)

	// Generate presigned URL for upload (expires in 1 hour)
	uploadURL, err := s.storage.GetUploadURL(context.Background(), tempKey, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &PresignedUploadResponse{
		UploadURL: uploadURL,
		ObjectKey: tempKey,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

// CompleteFileUpload finalizes file upload after successful upload to MinIO
func (s *FileService) CompleteFileUpload(userID, objectKey, filename, mimeType, fileHash string) (*models.UserFile, error) {
	ctx := context.Background()

	// Get file info from MinIO
	fileInfo, err := s.storage.GetFileInfo(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get or create FileHash record
	var fileHashRecord models.FileHash
	err = tx.Where("hash = ?", fileHash).First(&fileHashRecord).Error
	if err == gorm.ErrRecordNotFound {
		// New file, create hash record
		newObjectKey := fmt.Sprintf("files/%s", fileHash)

		// Move file to permanent location (in production, you might do this differently)
		fileHashRecord = models.FileHash{
			Hash:           fileHash,
			Size:           fileInfo.Size,
			MimeType:       mimeType,
			ReferenceCount: 1,
			MinIOKey:       newObjectKey,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		if err := tx.Create(&fileHashRecord).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create file hash record: %w", err)
		}
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to query file hash: %w", err)
	} else {
		// File already exists, increment reference count
		if err := tx.Model(&fileHashRecord).Update("reference_count", gorm.Expr("reference_count + 1")).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update reference count: %w", err)
		}
	}

	// Create UserFile record
	userFile := models.UserFile{
		ID:         uuid.New(),
		UserID:     userID,
		FileHash:   fileHash,
		Filename:   filename,
		IsPublic:   false,
		UploadedAt: time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := tx.Create(&userFile).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user file record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &userFile, nil
}

// GetUserFiles returns paginated list of user's files
func (s *FileService) GetUserFiles(userID string, offset, limit int) ([]UserFileResponse, int64, error) {
	var userFiles []models.UserFile
	var total int64

	// Count total files
	if err := s.db.Model(&models.UserFile{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user files: %w", err)
	}

	// Get files with file data
	err := s.db.Preload("FileData").
		Where("user_id = ?", userID).
		Order("uploaded_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&userFiles).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user files: %w", err)
	}

	// Convert to response format
	var response []UserFileResponse
	for _, file := range userFiles {
		response = append(response, UserFileResponse{
			ID:            file.ID,
			Filename:      file.Filename,
			Size:          file.FileData.Size,
			MimeType:      file.FileData.MimeType,
			IsPublic:      file.IsPublic,
			DownloadCount: file.DownloadCount,
			UploadedAt:    file.UploadedAt,
		})
	}

	return response, total, nil
}

// GetFileDownloadURL generates download URL for a file
func (s *FileService) GetFileDownloadURL(userID string, fileID uuid.UUID) (string, error) {
	var userFile models.UserFile

	query := s.db.Preload("FileData").Where("id = ?", fileID)

	// If not the file owner, only allow public files
	if userID != "" {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	} else {
		query = query.Where("is_public = ?", true)
	}

	err := query.First(&userFile).Error
	if err != nil {
		return "", fmt.Errorf("file not found or access denied: %w", err)
	}

	// Generate presigned download URL (expires in 1 hour)
	downloadURL, err := s.storage.GetFileURL(context.Background(), userFile.FileData.MinIOKey, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	// Increment download count
	go func() {
		s.db.Model(&userFile).Update("download_count", gorm.Expr("download_count + 1"))
	}()

	return downloadURL, nil
}

// DeleteUserFile deletes a user's file
func (s *FileService) DeleteUserFile(userID string, fileID uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get user file
	var userFile models.UserFile
	err := tx.Preload("FileData").Where("id = ? AND user_id = ?", fileID, userID).First(&userFile).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("file not found: %w", err)
	}

	// Delete user file record
	if err := tx.Delete(&userFile).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user file: %w", err)
	}

	// Decrement reference count
	var fileHash models.FileHash
	err = tx.Where("hash = ?", userFile.FileHash).First(&fileHash).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get file hash record: %w", err)
	}

	newRefCount := fileHash.ReferenceCount - 1
	if newRefCount <= 0 {
		// No more references, delete from storage and database
		if err := s.storage.DeleteFile(context.Background(), fileHash.MinIOKey); err != nil {
			// Log error but don't fail the transaction
			fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
		}

		if err := tx.Delete(&fileHash).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete file hash record: %w", err)
		}
	} else {
		// Update reference count
		if err := tx.Model(&fileHash).Update("reference_count", newRefCount).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reference count: %w", err)
		}
	}

	return tx.Commit().Error
}

// ToggleFilePublic toggles public/private status of a file
func (s *FileService) ToggleFilePublic(userID string, fileID uuid.UUID) error {
	err := s.db.Model(&models.UserFile{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Update("is_public", gorm.Expr("NOT is_public")).Error

	if err != nil {
		return fmt.Errorf("failed to toggle file public status: %w", err)
	}

	return nil
}

// GetPublicFileInfo gets public file info for sharing
func (s *FileService) GetPublicFileInfo(fileID uuid.UUID) (*PublicFileResponse, error) {
	var userFile models.UserFile

	err := s.db.Preload("FileData").Preload("User").
		Where("id = ? AND is_public = ?", fileID, true).
		First(&userFile).Error

	if err != nil {
		return nil, fmt.Errorf("public file not found: %w", err)
	}

	return &PublicFileResponse{
		ID:       fileID,
		Filename: userFile.Filename,
		Size:     userFile.FileData.Size,
		MimeType: userFile.FileData.MimeType,
	}, nil
}

// Response types
type PresignedUploadResponse struct {
	UploadURL string    `json:"upload_url"`
	ObjectKey string    `json:"object_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserFileResponse struct {
	ID            uuid.UUID `json:"id"`
	Filename      string    `json:"filename"`
	Size          int64     `json:"size"`
	MimeType      string    `json:"mime_type"`
	IsPublic      bool      `json:"is_public"`
	DownloadCount int       `json:"download_count"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

type PublicFileResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mime_type"`
}
