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

// GetProfile godoc
// @Summary Get user profile
// @Description Returns the current authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/profile [get]
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

// GetStorageInfo godoc
// @Summary Get storage information
// @Description Returns the current user's storage usage and quota information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Storage information"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/storage [get]
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

// GetStorageStatistics godoc
// @Summary Get storage statistics
// @Description Returns comprehensive storage statistics for the current user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Storage statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/storage/statistics [get]
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
