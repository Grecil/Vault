package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"filevault-backend/internal/errors"
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

// GenerateUploadURL godoc
// @Summary Generate upload URL
// @Description Generates a presigned URL for file upload
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{filename=string,size=int64,mime_type=string,file_hash=string} true "Upload request"
// @Success 200 {object} map[string]interface{} "Upload URL and metadata"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 402 {object} map[string]interface{} "Storage quota exceeded"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/upload-url [post]
func (h *FileHandler) GenerateUploadURL(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	var req struct {
		Filename string `json:"filename" binding:"required"`
		Size     int64  `json:"size" binding:"required"`
		MimeType string `json:"mime_type"`
		FileHash string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
		return
	}

	// Ensure user exists in database before checking quota
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserCreateFailed, "Failed to initialize user", err.Error()))
		return
	}

	// Check storage quota
	if err := h.userService.CheckStorageQuota(user.ID, req.Size); err != nil {
		c.JSON(http.StatusPaymentRequired, errors.ErrorResponse(errors.ErrStorageQuotaExceeded, err.Error()))
		return
	}

	response, err := h.fileService.GeneratePresignedUploadURL(user.ID, req.Filename, req.FileHash, req.Size, req.MimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileUploadFailed, "Failed to generate upload URL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteUpload godoc
// @Summary Complete file upload
// @Description Finalizes file upload after successful upload to storage
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{object_key=string,filename=string,mime_type=string,file_hash=string} true "Complete upload request"
// @Success 200 {object} map[string]interface{} "Upload completion confirmation"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/complete [post]
func (h *FileHandler) CompleteUpload(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	var req struct {
		ObjectKey string `json:"object_key" binding:"required"`
		Filename  string `json:"filename" binding:"required"`
		MimeType  string `json:"mime_type"`
		FileHash  string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
		return
	}

	userFile, err := h.fileService.CompleteFileUpload(user.ID, req.ObjectKey, req.Filename, req.MimeType, req.FileHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileUploadFailed, "Failed to complete upload", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file_id": userFile.ID,
	})
}

// ListFiles godoc
// @Summary List user files
// @Description Returns a paginated list of user's files
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} map[string]interface{} "List of files with pagination"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files [get]
func (h *FileHandler) ListFiles(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	// Ensure user exists in database
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserCreateFailed, "Failed to initialize user", err.Error()))
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
		c.JSON(http.StatusInternalServerError, errors.InternalServerErrorResponse("Failed to get files", err.Error()))
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

// DownloadFile godoc
// @Summary Download file
// @Description Generates a download URL for user's file
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "Download URL"
// @Failure 400 {object} map[string]interface{} "Invalid file ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Router /files/{id}/download [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	downloadURL, err := h.fileService.GetFileDownloadURL(user.ID, fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "File not found or access denied"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
	})
}

// DeleteFile godoc
// @Summary Delete file
// @Description Deletes a user's file
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "File deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid file ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	if err := h.fileService.DeleteUserFile(user.ID, fileID); err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "File not found or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileDeleteFailed, "Failed to delete file", err.Error()))
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// TogglePublic godoc
// @Summary Toggle file public status
// @Description Toggles file public status and manages share links
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "File public status updated"
// @Failure 400 {object} map[string]interface{} "Invalid file ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/{id}/public [patch]
func (h *FileHandler) TogglePublic(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	// First toggle the public status
	if err := h.fileService.ToggleFilePublic(user.ID, fileID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "File not found or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileToggleFailed, "Failed to toggle file public status", err.Error()))
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
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
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
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
		return
	}

	// Ensure user exists in database before checking quota
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserCreateFailed, "Failed to initialize user", err.Error()))
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
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileUploadFailed, "Failed to prepare batch upload", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// BatchCompleteUpload handles batch file upload completion
func (h *FileHandler) BatchCompleteUpload(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
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
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
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
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrFileUploadFailed, "Failed to complete batch upload", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicFile godoc
// @Summary Get public file info
// @Description Returns public file information
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "Public file information"
// @Failure 400 {object} map[string]interface{} "Invalid file ID"
// @Failure 404 {object} map[string]interface{} "Public file not found"
// @Router /public/files/{id} [get]
func (h *FileHandler) GetPublicFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	fileInfo, err := h.fileService.GetPublicFileInfo(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "Public file not found"))
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

// DownloadPublicFile godoc
// @Summary Download public file
// @Description Generates download URL for public file
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "Download URL"
// @Failure 400 {object} map[string]interface{} "Invalid file ID"
// @Failure 404 {object} map[string]interface{} "Public file not found"
// @Router /public/files/{id}/download [get]
func (h *FileHandler) DownloadPublicFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	downloadURL, err := h.fileService.GetFileDownloadURL("", fileID) // Empty userID for public access
	if err != nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "Public file not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
	})
}

// ShareFileDownload godoc
// @Summary Download file via share link
// @Description Handles file downloads via share links with tracking
// @Tags sharing
// @Param id path string true "Share ID"
// @Success 302 "Redirect to file download"
// @Failure 400 {object} map[string]interface{} "Invalid share ID"
// @Failure 404 {object} map[string]interface{} "Share link not found"
// @Router /share/{id} [get]
func (h *FileHandler) ShareFileDownload(c *gin.Context) {
	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidShareID, "Share ID required"))
		return
	}

	// Get file by share ID and increment download count
	userFile, err := h.fileService.GetFileByShareID(shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrFileNotFound, "Share link not found or file no longer available"))
		return
	}

	// Get actual MinIO URL for redirect
	downloadURL := h.fileService.GetPublicFileURL(userFile.FileData.MinIOKey)

	// Redirect to actual file with 302 (temporary redirect)
	c.Redirect(http.StatusFound, downloadURL)
}

// GetShareLink godoc
// @Summary Get share link
// @Description Returns the share link for a public file without toggling visibility
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "Share link"
// @Failure 400 {object} map[string]interface{} "Invalid file ID or file not public"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/{id}/share-link [get]
func (h *FileHandler) GetShareLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidFileID, "Invalid file ID"))
		return
	}

	// Verify file exists and is public
	files, _, err := h.fileService.GetUserFiles(user.ID, 0, 1000) // Get all files to find this one
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerErrorResponse("Failed to verify file", err.Error()))
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
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrFileAccessDenied, "File is not public"))
		return
	}

	// Get or create share link
	shareID, err := h.fileService.CreateOrGetShareLink(user.ID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrShareLinkFailed, "Failed to get share link", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_link": "/share/" + shareID,
	})
}
