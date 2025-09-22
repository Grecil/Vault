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
	// Check if file already exists (deduplication)
	var existingFileHash models.FileHash
	err := s.db.Where("hash = ?", fileHash).First(&existingFileHash).Error
	if err == nil {
		// File already exists, just create a UserFile record
		userFile := models.UserFile{
			ID:         uuid.New(),
			UserID:     userID,
			FileHash:   fileHash,
			Filename:   filename,
			IsPublic:   false,
			UploadedAt: time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}

		// Create UserFile record and increment reference count in a transaction
		tx := s.db.Begin()
		if err := tx.Create(&userFile).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create user file record for duplicate: %w", err)
		}

		if err := tx.Model(&existingFileHash).Update("reference_count", gorm.Expr("reference_count + 1")).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update reference count for duplicate: %w", err)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit duplicate file transaction: %w", err)
		}

		return &PresignedUploadResponse{
			UploadURL:    "", // No upload needed
			ObjectKey:    "",
			ExpiresAt:    time.Time{},
			IsDuplicate:  true,
			ExistingFile: &userFile,
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check for existing file: %w", err)
	}

	// File doesn't exist, generate upload URL directly to final location
	finalKey := fileHash // Simple hash-based key

	// Generate presigned URL for upload (expires in 1 hour)
	uploadURL, err := s.storage.GetUploadURL(context.Background(), finalKey, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &PresignedUploadResponse{
		UploadURL:   uploadURL,
		ObjectKey:   finalKey,
		ExpiresAt:   time.Now().Add(time.Hour),
		IsDuplicate: false,
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
		// New file, create hash record (file is already at final location)
		fileHashRecord = models.FileHash{
			Hash:           fileHash,
			Size:           fileInfo.Size,
			MimeType:       mimeType,
			ReferenceCount: 1,
			MinIOKey:       objectKey, // objectKey is already the final location: files/{hash}
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
		// File already exists - this shouldn't happen since we check for duplicates earlier
		// But if it does, increment reference count and clean up the duplicate upload
		if err := tx.Model(&fileHashRecord).Update("reference_count", gorm.Expr("reference_count + 1")).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update reference count: %w", err)
		}

		// Clean up the duplicate file that was just uploaded
		go func() {
			if err := s.storage.DeleteFile(context.Background(), objectKey); err != nil {
				// Log error but don't fail the operation since this is just cleanup
				fmt.Printf("Warning: failed to delete duplicate file %s: %v\n", objectKey, err)
			}
		}()
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
	response := make([]UserFileResponse, 0) // Initialize as empty slice, not nil
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

	var downloadURL string

	// For public files, return clean public URL; for private files, return presigned URL
	if userFile.IsPublic {
		// Return clean public URL (no auth parameters)
		downloadURL = s.storage.GetPublicFileURL(userFile.FileData.MinIOKey)
	} else {
		// Return presigned URL with short TTL for private files (1 minute)
		downloadURL, err = s.storage.GetFileURL(context.Background(), userFile.FileData.MinIOKey, time.Minute)
		if err != nil {
			return "", fmt.Errorf("failed to generate download URL: %w", err)
		}
	}

	// Increment download count
	go func() {
		s.db.Model(&userFile).Update("download_count", gorm.Expr("download_count + 1"))
	}()

	return downloadURL, nil
}

// DeleteUserFile deletes a user's file
func (s *FileService) DeleteUserFile(userID string, fileID uuid.UUID) error {
	fmt.Printf("UPDATED DELETION LOGIC: Starting deletion of file %s for user %s\n", fileID, userID)
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get user file without preloading FileData to avoid relation issues
	var userFile models.UserFile
	err := tx.Where("id = ? AND user_id = ?", fileID, userID).First(&userFile).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("database error finding file: %w", err)
	}

	// Get file hash record first (before deleting user file)
	var fileHash models.FileHash
	err = tx.Where("hash = ?", userFile.FileHash).First(&fileHash).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get file hash record: %w", err)
	}

	// Delete user file record first (hard delete to avoid foreign key issues)
	if err := tx.Unscoped().Delete(&userFile).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user file: %w", err)
	}

	// Check if there are any other user files still referencing this hash
	// Use Unscoped to count only non-soft-deleted records, but since we hard deleted above, this should be accurate
	var remainingRefs int64
	err = tx.Model(&models.UserFile{}).Where("file_hash = ?", userFile.FileHash).Count(&remainingRefs).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to count remaining file references: %w", err)
	}

	fmt.Printf("Remaining references for hash %s: %d\n", userFile.FileHash, remainingRefs)

	if remainingRefs == 0 {
		fmt.Printf("No more references, deleting file hash record for hash: %s\n", userFile.FileHash)

		// Clean up any orphaned soft-deleted records first
		fmt.Printf("Cleaning up orphaned soft-deleted records for hash: %s\n", userFile.FileHash)
		cleanupResult := tx.Unscoped().Where("file_hash = ? AND deleted_at IS NOT NULL", userFile.FileHash).Delete(&models.UserFile{})
		if cleanupResult.Error != nil {
			fmt.Printf("Warning: failed to cleanup soft-deleted records: %v\n", cleanupResult.Error)
		} else {
			fmt.Printf("Cleaned up %d orphaned soft-deleted records\n", cleanupResult.RowsAffected)
		}

		// No more references, delete from storage and database
		if err := s.storage.DeleteFile(context.Background(), fileHash.MinIOKey); err != nil {
			// Log error but don't fail the transaction - storage cleanup can be retried later
			fmt.Printf("Warning: failed to delete file from storage %s: %v\n", fileHash.MinIOKey, err)
		}

		if err := tx.Delete(&fileHash).Error; err != nil {
			tx.Rollback()
			fmt.Printf("ERROR: Failed to delete file hash record: %v\n", err)
			return fmt.Errorf("failed to delete file hash record: %w", err)
		}
		fmt.Printf("Successfully deleted file hash record for hash: %s\n", userFile.FileHash)
	} else {
		fmt.Printf("Still %d references remaining, updating reference count\n", remainingRefs)
		// Update reference count to match actual count
		if err := tx.Model(&fileHash).Update("reference_count", remainingRefs).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reference count: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	return nil
}

// ToggleFilePublic toggles public/private status of a file
func (s *FileService) ToggleFilePublic(userID string, fileID uuid.UUID) error {
	// Get file info with current status
	var userFile models.UserFile
	err := s.db.Preload("FileData").Where("id = ? AND user_id = ?", fileID, userID).First(&userFile).Error
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Calculate new public status
	newPublicStatus := !userFile.IsPublic

	// Start transaction for atomic update
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update database first
	err = tx.Model(&userFile).Update("is_public", newPublicStatus).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update database: %w", err)
	}

	// Update object tags in MinIO
	ctx := context.Background()
	if newPublicStatus {
		// Make public: set tag
		tags := map[string]string{"public": "true"}
		fmt.Printf("Setting public tag on object: %s with tags: %v\n", userFile.FileData.MinIOKey, tags)
		err = s.storage.SetObjectTags(ctx, userFile.FileData.MinIOKey, tags)
		if err != nil {
			fmt.Printf("Failed to set tags: %v\n", err)
		} else {
			fmt.Printf("Successfully set public tag on object: %s\n", userFile.FileData.MinIOKey)
		}
	} else {
		// Make private: remove tags
		fmt.Printf("Removing tags from object: %s\n", userFile.FileData.MinIOKey)
		err = s.storage.RemoveObjectTags(ctx, userFile.FileData.MinIOKey)
		if err != nil {
			fmt.Printf("Failed to remove tags: %v\n", err)
		} else {
			fmt.Printf("Successfully removed tags from object: %s\n", userFile.FileData.MinIOKey)
		}
	}

	if err != nil {
		// Revert database change if MinIO operation failed
		tx.Rollback()
		return fmt.Errorf("failed to update object access: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
	UploadURL    string           `json:"upload_url"`
	ObjectKey    string           `json:"object_key"`
	ExpiresAt    time.Time        `json:"expires_at"`
	IsDuplicate  bool             `json:"is_duplicate"`
	ExistingFile *models.UserFile `json:"existing_file,omitempty"`
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

// Batch upload types
type BatchFileRequest struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	FileHash string `json:"file_hash"`
}

type BatchFileResponse struct {
	FileHash     string      `json:"file_hash"`
	Status       string      `json:"status"` // "upload_required", "duplicate", "quota_exceeded"
	UploadID     string      `json:"upload_id,omitempty"`
	PresignedURL string      `json:"presigned_url,omitempty"`
	ExistingFile interface{} `json:"existing_file,omitempty"`
	Error        string      `json:"error,omitempty"`
}

type BatchPrepareResponse struct {
	BatchID    string              `json:"batch_id"`
	Files      []BatchFileResponse `json:"files"`
	QuotaCheck BatchQuotaCheck     `json:"quota_check"`
}

type BatchQuotaCheck struct {
	TotalSizeRequired int64 `json:"total_size_required"`
	QuotaAvailable    bool  `json:"quota_available"`
	QuotaExceeded     int64 `json:"quota_exceeded,omitempty"`
}

type BatchCompletedUpload struct {
	UploadID string `json:"upload_id"`
	FileHash string `json:"file_hash"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
}

type BatchCompleteResponse struct {
	BatchID        string        `json:"batch_id"`
	CompletedFiles []interface{} `json:"completed_files"`
	Errors         []string      `json:"errors,omitempty"`
}

// BatchPrepareUpload prepares multiple files for upload
func (s *FileService) BatchPrepareUpload(userID string, files []BatchFileRequest) (*BatchPrepareResponse, error) {
	batchID := uuid.New().String()

	// Calculate total size needed for new uploads
	var totalSizeRequired int64
	var duplicateHashes []string

	// Check for duplicates first
	fileHashes := make([]string, len(files))
	for i, file := range files {
		fileHashes[i] = file.FileHash
	}

	var existingHashes []models.FileHash
	s.db.Where("hash IN ?", fileHashes).Find(&existingHashes)

	existingHashMap := make(map[string]models.FileHash)
	for _, hash := range existingHashes {
		existingHashMap[hash.Hash] = hash
		duplicateHashes = append(duplicateHashes, hash.Hash)
	}

	// Calculate size for non-duplicates
	for _, file := range files {
		if _, isDuplicate := existingHashMap[file.FileHash]; !isDuplicate {
			totalSizeRequired += file.Size
		}
	}

	// Check quota for new uploads only
	quotaAvailable := true
	var quotaExceeded int64
	if totalSizeRequired > 0 {
		// Get current storage usage
		var currentUsage int64
		s.db.Model(&models.FileHash{}).
			Joins("JOIN user_files ON file_hashes.hash = user_files.file_hash").
			Where("user_files.user_id = ?", userID).
			Select("COALESCE(SUM(file_hashes.size), 0)").
			Scan(&currentUsage)

		const maxStorage = 10 * 1024 * 1024 * 1024 // 10GB
		if currentUsage+totalSizeRequired > maxStorage {
			quotaAvailable = false
			quotaExceeded = (currentUsage + totalSizeRequired) - maxStorage
		}
	}

	// Prepare response for each file
	var fileResponses []BatchFileResponse

	for _, file := range files {
		if existingHash, isDuplicate := existingHashMap[file.FileHash]; isDuplicate {
			// File is duplicate - create UserFile record
			userFile := models.UserFile{
				ID:         uuid.New(),
				UserID:     userID,
				FileHash:   file.FileHash,
				Filename:   file.Filename,
				IsPublic:   false,
				UploadedAt: time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			}

			// Create UserFile record in transaction
			tx := s.db.Begin()
			if err := tx.Create(&userFile).Error; err != nil {
				tx.Rollback()
				fileResponses = append(fileResponses, BatchFileResponse{
					FileHash: file.FileHash,
					Status:   "error",
					Error:    "Failed to link duplicate file",
				})
				continue
			}

			// Increment reference count
			if err := tx.Model(&existingHash).Update("reference_count", gorm.Expr("reference_count + 1")).Error; err != nil {
				tx.Rollback()
				fileResponses = append(fileResponses, BatchFileResponse{
					FileHash: file.FileHash,
					Status:   "error",
					Error:    "Failed to update reference count",
				})
				continue
			}

			tx.Commit()

			fileResponses = append(fileResponses, BatchFileResponse{
				FileHash: file.FileHash,
				Status:   "duplicate",
				ExistingFile: map[string]interface{}{
					"id":       userFile.ID,
					"filename": file.Filename,
					"size":     existingHash.Size,
				},
			})
		} else if !quotaAvailable {
			// Quota exceeded
			fileResponses = append(fileResponses, BatchFileResponse{
				FileHash: file.FileHash,
				Status:   "quota_exceeded",
				Error:    "Storage quota would be exceeded",
			})
		} else {
			// Generate upload URL
			uploadID := uuid.New().String()
			objectKey := fmt.Sprintf("uploads/%s/%s", userID, uploadID)

			presignedURL, err := s.storage.GetUploadURL(context.Background(), objectKey, 15*time.Minute)
			if err != nil {
				fileResponses = append(fileResponses, BatchFileResponse{
					FileHash: file.FileHash,
					Status:   "error",
					Error:    "Failed to generate upload URL",
				})
				continue
			}

			fileResponses = append(fileResponses, BatchFileResponse{
				FileHash:     file.FileHash,
				Status:       "upload_required",
				UploadID:     uploadID,
				PresignedURL: presignedURL,
			})
		}
	}

	return &BatchPrepareResponse{
		BatchID: batchID,
		Files:   fileResponses,
		QuotaCheck: BatchQuotaCheck{
			TotalSizeRequired: totalSizeRequired,
			QuotaAvailable:    quotaAvailable,
			QuotaExceeded:     quotaExceeded,
		},
	}, nil
}

// BatchCompleteUpload completes multiple file uploads
func (s *FileService) BatchCompleteUpload(userID, batchID string, completedUploads []BatchCompletedUpload) (*BatchCompleteResponse, error) {
	var completedFiles []interface{}
	var errors []string

	for _, upload := range completedUploads {
		objectKey := fmt.Sprintf("uploads/%s/%s", userID, upload.UploadID)

		// Complete individual file upload
		userFile, err := s.CompleteFileUpload(userID, objectKey, upload.Filename, upload.MimeType, upload.FileHash)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to complete upload for %s: %v", upload.Filename, err))
			continue
		}

		completedFiles = append(completedFiles, map[string]interface{}{
			"id":       userFile.ID,
			"filename": userFile.Filename,
			"size":     userFile.FileHash,
		})
	}

	return &BatchCompleteResponse{
		BatchID:        batchID,
		CompletedFiles: completedFiles,
		Errors:         errors,
	}, nil
}
