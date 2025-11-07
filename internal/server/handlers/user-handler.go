package handlers

import (
	"AuthServer/internal/domain/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func getUserFromDatabase(c *gin.Context) (*models.User, string) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "need to include a Authorization header"})
		return nil, ""
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := tokenService.DecodeAccessToken(tokenString)
	userID, ok := claims["user-id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "debug": claims["user-id"].(string)})
		return nil, ""
	}

	user, err := userService.FindById(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return nil, ""
	}

	return user, userID
}

func (s *Server) GetUserData(c *gin.Context) {
	user, _ := getUserFromDatabase(c)

	if user == nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"full_name":  user.FullName,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})

	return
}
