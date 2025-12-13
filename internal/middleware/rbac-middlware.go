package middleware

import (
	"AuthServer/internal/domain/roles"
	"AuthServer/internal/service"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireRole(rbacService *service.RBACService, tokenService service.ITokenService, requiredRole roles.Role, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenService.DecodeAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		userID, ok := claims["user-id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		var resourceID *string
		if resourceIDParam != "" {
			rid := c.Param(resourceIDParam)
			if rid == "" {
				rid = c.Query(resourceIDParam)
			}
			if rid != "" {
				resourceID = &rid
			}
		}

		hasPermission, err := rbacService.HasPermission(userID, requiredRole, resourceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func RequireAnyRole(rbacService *service.RBACService, tokenService service.ITokenService, requiredRoles []roles.Role, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenService.DecodeAccessToken(tokenString)
		if err != nil {
			log.Printf("Token decode error: %v", err) // DEBUG
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		userID, ok := claims["user-id"].(string)
		if !ok {
			log.Printf("Claims user-id not found: %+v", claims) // DEBUG
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		log.Printf("User ID from token: %s", userID) // DEBUG

		var resourceID *string
		if resourceIDParam != "" {
			rid := c.Param(resourceIDParam)
			if rid == "" {
				rid = c.Query(resourceIDParam)
			}
			if rid != "" {
				resourceID = &rid
			}
		}

		hasPermission := false
		for _, role := range requiredRoles {
			log.Printf("Checking if user %s has role %s", userID, role) // DEBUG
			permitted, err := rbacService.HasPermission(userID, role, resourceID)
			if err != nil {
				log.Printf("Permission check error: %v", err) // DEBUG
			}
			log.Printf("Has permission: %v", permitted) // DEBUG
			if err == nil && permitted {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			log.Printf("User %s does not have required roles: %v", userID, requiredRoles) // DEBUG
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
