package handlers

import (
	"net/http"

	"filevault-backend/internal/errors"
	"filevault-backend/internal/middleware"
	"filevault-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	// Get or create user in database
	dbUser, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserCreateFailed, "Failed to get user profile", err.Error()))
		return
	}

	// Update user role in context
	user.Role = dbUser.Role

	c.JSON(http.StatusOK, gin.H{
		"id":            dbUser.ID,
		"email":         user.Email,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"role":          dbUser.Role,
		"storage_quota": dbUser.StorageQuota,
		"storage_used":  dbUser.StorageUsed,
		"created_at":    dbUser.CreatedAt,
	})
}

// GetStorageInfo returns user's storage usage information
func (h *UserHandler) GetStorageInfo(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	used, quota, err := h.userService.GetUserStorageInfo(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrStorageInfoFailed, "Failed to get storage info", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"storage_used":  used,
		"storage_quota": quota,
		"storage_free":  quota - used,
		"usage_percent": float64(used) / float64(quota) * 100,
	})
}

// GetStorageStatistics returns comprehensive storage statistics for the user
func (h *UserHandler) GetStorageStatistics(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, errors.UnauthorizedResponse("User not found"))
		return
	}

	// Ensure user exists in database
	_, err := h.userService.GetOrCreateUser(user.ID, user.Email, user.FirstName, user.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrUserCreateFailed, "Failed to get user profile", err.Error()))
		return
	}

	statistics, err := h.userService.GetStorageStatistics(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse(errors.ErrStorageStatsFailed, "Failed to get storage statistics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, statistics)
}
