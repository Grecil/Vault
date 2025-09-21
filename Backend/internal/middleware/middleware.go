package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"filevault-backend/internal/config"
	"filevault-backend/internal/models"
	"filevault-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ClerkClaims struct {
	Sub       string `json:"sub"`   // User ID
	Email     string `json:"email"` // User email
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

type AuthenticatedUser struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Role      models.UserRole
}

const UserContextKey = "user"

// CORS middleware
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// RequireAuth middleware validates Clerk JWT tokens
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse and validate JWT token
		claims := &ClerkClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// For now, we'll use a simple validation
			// In production, you'd fetch Clerk's public keys from their JWKS endpoint
			return []byte(cfg.ClerkSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Create authenticated user context
		user := &AuthenticatedUser{
			ID:        claims.Sub,
			Email:     claims.Email,
			FirstName: claims.FirstName,
			LastName:  claims.LastName,
			Role:      models.UserRoleUser, // Default role, will be updated from DB
		}

		c.Set(UserContextKey, user)
		c.Next()
	})
}

// RequireAdmin middleware requires admin role
func RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user := GetUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if user.Role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	})
}

// OptionalAuth middleware validates auth if present but doesn't require it
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Use the same logic as RequireAuth but don't abort on failure
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		claims := &ClerkClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.ClerkSecretKey), nil
		})

		if err == nil && token.Valid {
			user := &AuthenticatedUser{
				ID:        claims.Sub,
				Email:     claims.Email,
				FirstName: claims.FirstName,
				LastName:  claims.LastName,
				Role:      models.UserRoleUser,
			}
			c.Set(UserContextKey, user)
		}

		c.Next()
	})
}

// GetUserFromContext extracts the authenticated user from gin context
func GetUserFromContext(c *gin.Context) *AuthenticatedUser {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}

	authUser, ok := user.(*AuthenticatedUser)
	if !ok {
		return nil
	}

	return authUser
}

// RequestLogger middleware for structured logging
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logData := map[string]interface{}{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency.String(),
			"client_ip":   param.ClientIP,
			"user_agent":  param.Request.UserAgent(),
		}

		if param.ErrorMessage != "" {
			logData["error"] = param.ErrorMessage
		}

		jsonLog, _ := json.Marshal(logData)
		return string(jsonLog) + "\n"
	})
}

// RateLimit middleware - simple unified rate limiting
func RateLimit(rateLimitService *services.RateLimitService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get identifier (user ID or IP)
		user := GetUserFromContext(c)
		identifier := c.ClientIP() // Default to IP
		if user != nil {
			identifier = user.ID // Use user ID if authenticated
		}

		result := rateLimitService.CheckRateLimit(identifier)

		// Set headers
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))

		if !result.Allowed {
			retryAfter := time.Until(result.ResetTime).Seconds()
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests. Please slow down.",
				"message": "You are making requests too quickly. Please wait before trying again.",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}
