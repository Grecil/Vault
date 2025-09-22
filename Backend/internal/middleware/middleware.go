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

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// ClerkJWKSClient stores the JWKS client for token verification
var ClerkJWKSClient *jwks.Client

// InitializeClerk sets up the Clerk SDK with the secret key
func InitializeClerk(cfg *config.Config) {
	clerk.SetKey(cfg.ClerkSecretKey)

	// Initialize JWKS client for manual verification
	config := &clerk.ClientConfig{}
	config.Key = clerk.String(cfg.ClerkSecretKey)
	ClerkJWKSClient = jwks.NewClient(config)
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

// RequireAuth middleware validates Clerk JWT tokens using proper verification
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get the session token from Authorization header or __session cookie
		sessionToken := getSessionToken(c.Request)
		if sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Decode the session JWT to find the key ID
		unsafeClaims, err := jwt.Decode(c.Request.Context(), &jwt.DecodeParams{
			Token: sessionToken,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Fetch the JSON Web Key
		jwk, err := jwt.GetJSONWebKey(c.Request.Context(), &jwt.GetJSONWebKeyParams{
			KeyID:      unsafeClaims.KeyID,
			JWKSClient: ClerkJWKSClient,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			c.Abort()
			return
		}

		// Verify the session with 1 minute leeway for clock skew
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token:  sessionToken,
			JWK:    jwk,
			Leeway: time.Minute, // 1 minute leeway for clock skew
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
			c.Abort()
			return
		}

		// Create authenticated user context
		// Note: We'll get user details from the database or user API if needed
		user := &AuthenticatedUser{
			ID:        claims.Subject,
			Email:     "", // We'll fetch this from Clerk User API if needed
			FirstName: "",
			LastName:  "",
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
		// Get the session token from Authorization header or __session cookie
		sessionToken := getSessionToken(c.Request)
		if sessionToken == "" {
			c.Next()
			return
		}

		// Try to verify token, but don't abort on failure
		unsafeClaims, err := jwt.Decode(c.Request.Context(), &jwt.DecodeParams{
			Token: sessionToken,
		})
		if err != nil {
			c.Next()
			return
		}

		jwk, err := jwt.GetJSONWebKey(c.Request.Context(), &jwt.GetJSONWebKeyParams{
			KeyID:      unsafeClaims.KeyID,
			JWKSClient: ClerkJWKSClient,
		})
		if err != nil {
			c.Next()
			return
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token:  sessionToken,
			JWK:    jwk,
			Leeway: time.Minute, // 1 minute leeway for clock skew
		})
		if err != nil {
			c.Next()
			return
		}

		user := &AuthenticatedUser{
			ID:        claims.Subject,
			Email:     "", // We'll fetch this from Clerk User API if needed
			FirstName: "",
			LastName:  "",
			Role:      models.UserRoleUser,
		}
		c.Set(UserContextKey, user)

		c.Next()
	})
}

// getSessionToken retrieves the session token from either the Authorization header
// or the __session cookie, depending on the request type
func getSessionToken(r *http.Request) string {
	// First try to get the token from the Authorization header (cross-origin)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// If not found in header, try to get from __session cookie (same-origin)
	cookie, err := r.Cookie("__session")
	if err != nil {
		return ""
	}
	return cookie.Value
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
