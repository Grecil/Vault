package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"filevault-backend/internal/errors"
	"filevault-backend/internal/models"
	"filevault-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService *services.UserService
	fileService *services.FileService
}

func NewAdminHandler(userService *services.UserService, fileService *services.FileService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		fileService: fileService,
	}
}

// ListUsers godoc
// @Summary List all users (Admin only)
// @Description Returns a paginated list of all users
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50) maximum(100)
// @Success 200 {object} map[string]interface{} "List of users with pagination"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max 100 items per page
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.ListUsers(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerErrorResponse("Failed to get users", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// DeleteUser godoc
// @Summary Delete user (Admin only)
// @Description Deletes a user from the system
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("User ID required"))
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, errors.ErrorResponse(errors.ErrUserNotFound, "User not found"))
		} else {
			c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserDeleteFailed, "Failed to delete user", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// UpdateUserRole godoc
// @Summary Update user role (Admin only)
// @Description Updates a user's role (user or admin)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body object{role=string} true "Role update request"
// @Success 200 {object} map[string]interface{} "User role updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users/{id}/role [patch]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("User ID required"))
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
		return
	}

	// Validate role
	var role models.UserRole
	switch req.Role {
	case "user":
		role = models.UserRoleUser
	case "admin":
		role = models.UserRoleAdmin
	default:
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidRole, "Invalid role. Must be 'user' or 'admin'"))
		return
	}

	if err := h.userService.UpdateUserRole(userID, role); err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserUpdateFailed, "Failed to update user role", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
	})
}

// UpdateUserQuota godoc
// @Summary Update user storage quota (Admin only)
// @Description Updates a user's storage quota in MB
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body object{quota_mb=int64} true "Quota update request"
// @Success 200 {object} map[string]interface{} "User storage quota updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users/{id}/quota [patch]
func (h *AdminHandler) UpdateUserQuota(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("User ID required"))
		return
	}

	var req struct {
		QuotaMB int64 `json:"quota_mb" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ValidationErrorResponse("Invalid request body", err.Error()))
		return
	}

	if req.QuotaMB <= 0 {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse(errors.ErrInvalidQuota, "Quota must be greater than 0"))
		return
	}

	if err := h.userService.UpdateStorageQuota(userID, req.QuotaMB); err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserUpdateFailed, "Failed to update storage quota", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User storage quota updated successfully",
		"quota_mb": req.QuotaMB,
	})
}

// GetStats godoc
// @Summary Get system statistics (Admin only)
// @Description Returns system-wide statistics
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "System statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Router /admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	// For now, return basic stats
	// In the future, you could implement more detailed analytics
	c.JSON(http.StatusOK, gin.H{
		"message": "Stats endpoint - coming soon",
		"stats": gin.H{
			"total_users":        0,
			"total_files":        0,
			"total_storage_used": 0,
		},
	})
}
