package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"filevault-backend/internal/middleware"
	"filevault-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileService *services.FileService
	userService *services.UserService
}

func NewFileHandler(fileService *services.FileService, userService *services.UserService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		userService: userService,
	}
}

// GenerateUploadURL generates presigned URL for file upload
func (h *FileHandler) GenerateUploadURL(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Filename string `json:"filename" binding:"required"`
		Size     int64  `json:"size" binding:"required"`
		MimeType string `json:"mime_type"`
		FileHash string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user exists in database before checking quota
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize user"})
		return
	}

	// Check storage quota
	if err := h.userService.CheckStorageQuota(user.ID, req.Size); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error": err.Error(),
			"code":  "STORAGE_QUOTA_EXCEEDED",
		})
		return
	}

	response, err := h.fileService.GeneratePresignedUploadURL(user.ID, req.Filename, req.FileHash, req.Size, req.MimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteUpload finalizes file upload
func (h *FileHandler) CompleteUpload(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		ObjectKey string `json:"object_key" binding:"required"`
		Filename  string `json:"filename" binding:"required"`
		MimeType  string `json:"mime_type"`
		FileHash  string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userFile, err := h.fileService.CompleteFileUpload(user.ID, req.ObjectKey, req.Filename, req.MimeType, req.FileHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file_id": userFile.ID,
	})
}

// ListFiles returns user's files
func (h *FileHandler) ListFiles(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Ensure user exists in database
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize user"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max 100 items per page
	}

	offset := (page - 1) * limit

	files, total, err := h.fileService.GetUserFiles(user.ID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files":    files,
		"total":    total,
		"has_more": int64(offset+limit) < total,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// DownloadFile generates download URL for user's file
func (h *FileHandler) DownloadFile(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	downloadURL, err := h.fileService.GetFileDownloadURL(user.ID, fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found or access denied",
			"code":  "FILE_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
	})
}

// DeleteFile deletes user's file
func (h *FileHandler) DeleteFile(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	fmt.Printf("Attempting to delete file %s for user %s\n", fileID, user.ID)

	if err := h.fileService.DeleteUserFile(user.ID, fileID); err != nil {
		fmt.Printf("Error deleting file %s: %v\n", fileID, err)
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found or access denied",
				"code":  "FILE_NOT_FOUND",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete file",
				"code":    "DELETE_FAILED",
				"details": err.Error(),
			})
		}
		return
	}

	fmt.Printf("Successfully deleted file %s\n", fileID)
	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// TogglePublic toggles file public status and manages share links
func (h *FileHandler) TogglePublic(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// First toggle the public status
	if err := h.fileService.ToggleFilePublic(user.ID, fileID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found or access denied",
				"code":  "FILE_NOT_FOUND",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to toggle file public status",
				"code":  "TOGGLE_FAILED",
			})
		}
		return
	}

	// Check if file is now public and create/get share link
	var shareLink string
	var isPublic bool

	// Get updated file status
	files, _, err := h.fileService.GetUserFiles(user.ID, 0, 1000) // Get all files to find this one
	if err == nil {
		for _, file := range files {
			if file.ID.String() == fileID.String() {
				isPublic = file.IsPublic
				break
			}
		}
	}

	if isPublic {
		// File is now public, create or get share link
		shareID, err := h.fileService.CreateOrGetShareLink(user.ID, fileID)
		if err == nil {
			shareLink = "/share/" + shareID
		}
		// If share link creation fails, we still return success for the toggle
	} else {
		// File is now private, delete share link
		h.fileService.DeleteShareLink(user.ID, fileID)
	}

	response := gin.H{
		"message":   "File public status updated",
		"is_public": isPublic,
	}

	if shareLink != "" {
		response["share_link"] = shareLink
	}

	c.JSON(http.StatusOK, response)
}

// BatchPrepareUpload handles batch file upload preparation
func (h *FileHandler) BatchPrepareUpload(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Files []struct {
			Filename string `json:"filename" binding:"required"`
			Size     int64  `json:"size" binding:"required"`
			MimeType string `json:"mime_type"`
			FileHash string `json:"file_hash" binding:"required"`
		} `json:"files" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user exists in database before checking quota
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize user"})
		return
	}

	// Convert request struct to service struct
	files := make([]services.BatchFileRequest, len(req.Files))
	for i, f := range req.Files {
		files[i] = services.BatchFileRequest{
			Filename: f.Filename,
			Size:     f.Size,
			MimeType: f.MimeType,
			FileHash: f.FileHash,
		}
	}

	response, err := h.fileService.BatchPrepareUpload(user.ID, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BatchCompleteUpload handles batch file upload completion
func (h *FileHandler) BatchCompleteUpload(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		BatchID          string `json:"batch_id" binding:"required"`
		CompletedUploads []struct {
			UploadID string `json:"upload_id" binding:"required"`
			FileHash string `json:"file_hash" binding:"required"`
			Filename string `json:"filename" binding:"required"`
			MimeType string `json:"mime_type"`
		} `json:"completed_uploads" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request struct to service struct
	completedUploads := make([]services.BatchCompletedUpload, len(req.CompletedUploads))
	for i, upload := range req.CompletedUploads {
		completedUploads[i] = services.BatchCompletedUpload{
			UploadID: upload.UploadID,
			FileHash: upload.FileHash,
			Filename: upload.Filename,
			MimeType: upload.MimeType,
		}
	}

	response, err := h.fileService.BatchCompleteUpload(user.ID, req.BatchID, completedUploads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicFile returns public file info
func (h *FileHandler) GetPublicFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	fileInfo, err := h.fileService.GetPublicFileInfo(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public file not found"})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

// DownloadPublicFile generates download URL for public file
func (h *FileHandler) DownloadPublicFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	downloadURL, err := h.fileService.GetFileDownloadURL("", fileID) // Empty userID for public access
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
	})
}

// ShareFileDownload handles file downloads via share links with tracking
func (h *FileHandler) ShareFileDownload(c *gin.Context) {
	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share ID required"})
		return
	}

	// Get file by share ID and increment download count
	userFile, err := h.fileService.GetFileByShareID(shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share link not found or file no longer available"})
		return
	}

	// Get actual MinIO URL for redirect
	downloadURL := h.fileService.GetPublicFileURL(userFile.FileData.MinIOKey)

	// Redirect to actual file with 302 (temporary redirect)
	c.Redirect(http.StatusFound, downloadURL)
}

// GetShareLink returns the share link for a public file without toggling visibility
func (h *FileHandler) GetShareLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Verify file exists and is public
	files, _, err := h.fileService.GetUserFiles(user.ID, 0, 1000) // Get all files to find this one
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify file"})
		return
	}

	var isPublic bool
	for _, file := range files {
		if file.ID.String() == fileID.String() {
			isPublic = file.IsPublic
			break
		}
	}

	if !isPublic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not public"})
		return
	}

	// Get or create share link
	shareID, err := h.fileService.CreateOrGetShareLink(user.ID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get share link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_link": "/share/" + shareID,
	})
}
