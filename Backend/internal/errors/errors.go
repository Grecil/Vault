package errors

import "github.com/gin-gonic/gin"

// Error codes for consistent API responses
const (
	// Authentication & Authorization errors
	ErrAuthRequired            = "AUTH_REQUIRED"
	ErrInvalidToken            = "INVALID_TOKEN"
	ErrTokenVerificationFailed = "TOKEN_VERIFICATION_FAILED"
	ErrInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrAdminAccessRequired     = "ADMIN_ACCESS_REQUIRED"

	// User-related errors
	ErrUserNotFound     = "USER_NOT_FOUND"
	ErrUserCreateFailed = "USER_CREATE_FAILED"
	ErrUserDeleteFailed = "USER_DELETE_FAILED"
	ErrUserUpdateFailed = "USER_UPDATE_FAILED"

	// File-related errors
	ErrFileNotFound     = "FILE_NOT_FOUND"
	ErrFileUploadFailed = "FILE_UPLOAD_FAILED"
	ErrFileDeleteFailed = "FILE_DELETE_FAILED"
	ErrFileAccessDenied = "FILE_ACCESS_DENIED"
	ErrFileToggleFailed = "FILE_TOGGLE_FAILED"
	ErrShareLinkFailed  = "SHARE_LINK_FAILED"
	ErrInvalidFileID    = "INVALID_FILE_ID"
	ErrInvalidShareID   = "INVALID_SHARE_ID"

	// Storage-related errors
	ErrStorageQuotaExceeded = "STORAGE_QUOTA_EXCEEDED"
	ErrStorageInfoFailed    = "STORAGE_INFO_FAILED"
	ErrStorageStatsFailed   = "STORAGE_STATS_FAILED"

	// Validation errors
	ErrInvalidInput     = "INVALID_INPUT"
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrInvalidRole      = "INVALID_ROLE"
	ErrInvalidQuota     = "INVALID_QUOTA"
	ErrRequiredField    = "REQUIRED_FIELD"

	// Rate limiting errors
	ErrRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	// Server errors
	ErrInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrDatabaseError      = "DATABASE_ERROR"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// ErrorResponse creates a standardized error response
func ErrorResponse(code string, message string, details ...string) gin.H {
	response := gin.H{
		"error": message,
		"code":  code,
	}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	return response
}

// ValidationErrorResponse creates a standardized validation error response
func ValidationErrorResponse(message string, details ...string) gin.H {
	return ErrorResponse(ErrValidationFailed, message, details...)
}

// NotFoundResponse creates a standardized not found error response
func NotFoundResponse(resource string) gin.H {
	return ErrorResponse(ErrFileNotFound, resource+" not found")
}

// UnauthorizedResponse creates a standardized unauthorized error response
func UnauthorizedResponse(message string) gin.H {
	return ErrorResponse(ErrAuthRequired, message)
}

// ForbiddenResponse creates a standardized forbidden error response
func ForbiddenResponse(message string) gin.H {
	return ErrorResponse(ErrInsufficientPermissions, message)
}

// InternalServerErrorResponse creates a standardized internal server error response
func InternalServerErrorResponse(message string, details ...string) gin.H {
	return ErrorResponse(ErrInternalServer, message, details...)
}
