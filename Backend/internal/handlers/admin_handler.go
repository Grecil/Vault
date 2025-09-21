package handlers

import (
	"net/http"
	"strconv"
	"strings"

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

// ListUsers returns paginated list of all users
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
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

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
				"code":  "USER_NOT_FOUND",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user",
				"code":  "DELETE_FAILED",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// UpdateUserRole updates a user's role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'user' or 'admin'"})
		return
	}

	if err := h.userService.UpdateUserRole(userID, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
	})
}

// UpdateUserQuota updates a user's storage quota
func (h *AdminHandler) UpdateUserQuota(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	var req struct {
		QuotaMB int64 `json:"quota_mb" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.QuotaMB <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quota must be greater than 0"})
		return
	}

	if err := h.userService.UpdateStorageQuota(userID, req.QuotaMB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User storage quota updated successfully",
		"quota_mb": req.QuotaMB,
	})
}

// GetStats returns system statistics
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
