package handlers

import (
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
		"files": files,
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

	if err := h.fileService.DeleteUserFile(user.ID, fileID); err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found or access denied",
				"code":  "FILE_NOT_FOUND",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file",
				"code":  "DELETE_FAILED",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// TogglePublic toggles file public status
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

	c.JSON(http.StatusOK, gin.H{
		"message": "File public status updated",
	})
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
