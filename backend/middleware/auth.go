package middleware

import (
	"net/http"
	"strings"

	"inventory-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header required", "")
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Invalid authorization header format. Use: Bearer <token>", "")
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			utils.Error(c, http.StatusUnauthorized, "Token is required", "")
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// Check error type for better message
			errMsg := "Invalid or expired token"
			if strings.Contains(err.Error(), "expired") {
				errMsg = "Token has expired"
			} else if strings.Contains(err.Error(), "signature") {
				errMsg = "Invalid token signature"
			}
			utils.Error(c, http.StatusUnauthorized, errMsg, err.Error())
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("tokenClaims", claims)

		c.Next()
	}
}

// RoleMiddleware checks if user has one of the allowed roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		role, exists := c.Get("role")
		if !exists {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			utils.Error(c, http.StatusInternalServerError, "Invalid role format", "")
			c.Abort()
			return
		}

		// Check if role is allowed
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		// Role not allowed
		utils.Error(c, http.StatusForbidden, 
			"Access denied. Required role: "+strings.Join(allowedRoles, ", "),
			"User role: "+roleStr,
		)
		c.Abort()
	}
}

// OptionalAuthMiddleware validates token if present but doesn't require it
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("tokenClaims", claims)
		}

		c.Next()
	}
}

// PermissionMiddleware checks for specific permissions
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		role, exists := c.Get("role")
		if !exists {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated", "")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			utils.Error(c, http.StatusInternalServerError, "Invalid role format", "")
			c.Abort()
			return
		}

		// Define permissions for each role
		permissions := map[string][]string{
			"superadmin": {
				"users:read", "users:write", "users:delete",
				"products:read", "products:write", "products:delete",
				"transactions:read", "transactions:write", "transactions:delete",
				"reports:read", "reports:write",
				"settings:read", "settings:write",
			},
			"head": {
				"users:read",
				"products:read", "products:write",
				"transactions:read", "transactions:write",
				"reports:read",
			},
			"produksi": {
				"products:read",
				"transactions:read", "transactions:write",
			},
		}

		// Check if role has the required permission
		rolePermissions, exists := permissions[roleStr]
		if !exists {
			utils.Error(c, http.StatusForbidden, "Invalid role", "")
			c.Abort()
			return
		}

		for _, perm := range rolePermissions {
			if perm == requiredPermission {
				c.Next()
				return
			}
		}

		utils.Error(c, http.StatusForbidden, 
			"Insufficient permissions. Required: "+requiredPermission,
			"User role: "+roleStr,
		)
		c.Abort()
	}
}

// GetUserID returns user ID from context
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int)
	return id, ok
}

// GetUserRole returns user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}

// GetUsername returns username from context
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	usernameStr, ok := username.(string)
	return usernameStr, ok
}