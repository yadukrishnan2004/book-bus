package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"book-bus/internal/domain"
)

const (
	// Context keys
	CtxUserIDKey = "userID"
	CtxEmailKey  = "userEmail"
	CtxRoleKey   = "userRole"
)

// JWTAuth verifies Bearer tokens in Authorization header and populates Gin context.
func JWTAuth(authSvc domain.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		userID, email, role, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, userID)
		c.Set(CtxEmailKey, email)
		c.Set(CtxRoleKey, role)
		c.Next()
	}
}

// OptionalJWTAuth extracts and validates JWT token if present, but does not reject unauthenticated requests.
func OptionalJWTAuth(authSvc domain.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr != "" {
			userID, email, role, err := authSvc.ValidateToken(tokenStr)
			if err == nil {
				c.Set(CtxUserIDKey, userID)
				c.Set(CtxEmailKey, email)
				c.Set(CtxRoleKey, role)
			}
		}
		c.Next()
	}
}

// RequireRole enforces role-based access control (RBAC).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowedMap := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedMap[r] = true
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxRoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		role, ok := roleVal.(string)
		if !ok || !allowedMap[role] {
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from Gin context if set.
func GetUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(CtxUserIDKey)
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// extractBearerToken gets token from Authorization: Bearer <token> header.
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
