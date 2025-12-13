package handlers

import (
	"AuthServer/internal/domain/roles"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============= ROLE ASSIGNMENT =============

func (s *Server) AssignGlobalRole(c *gin.Context) {
	assignerID, _ := c.Get("user_id")

	var input struct {
		UserID string     `json:"user_id" binding:"required"`
		Role   roles.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate hierarchical role
	if _, ok := roles.RoleHierarchy[input.Role]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a valid global hierarchical role"})
		return
	}

	err := rbacService.AssignRole(input.UserID, input.Role, nil, nil, assignerID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "global role assigned",
		"user_id": input.UserID,
		"role":    input.Role,
	})
}

func (s *Server) AssignResourceRole(c *gin.Context) {
	assignerID, _ := c.Get("user_id")

	var input struct {
		UserID     string     `json:"user_id" binding:"required"`
		Role       roles.Role `json:"role" binding:"required"`
		ResourceID string     `json:"resource_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Role != roles.RoleProjectEditor && input.Role != roles.RoleProjectViewer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a valid resource-specific role"})
		return
	}

	err := rbacService.AssignRole(input.UserID, input.Role, &input.ResourceID, nil, assignerID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "resource role assigned",
		"user_id":     input.UserID,
		"role":        input.Role,
		"resource_id": input.ResourceID,
	})
}

func (s *Server) AssignTemporaryRole(c *gin.Context) {
	assignerID, _ := c.Get("user_id")

	var input struct {
		UserID          string     `json:"user_id" binding:"required"`
		Role            roles.Role `json:"role" binding:"required"`
		ResourceID      *string    `json:"resource_id"`
		DurationMinutes int        `json:"duration_minutes" binding:"required,min=1,max=1440"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiresAt := time.Now().Add(time.Duration(input.DurationMinutes) * time.Minute)

	err := rbacService.AssignRole(input.UserID, input.Role, input.ResourceID, &expiresAt, assignerID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "temporary role assigned",
		"user_id":     input.UserID,
		"role":        input.Role,
		"resource_id": input.ResourceID,
		"expires_at":  expiresAt,
	})
}

func (s *Server) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}

	userRoles, err := rbacService.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   userRoles,
	})
}

func (s *Server) GetMyRoles(c *gin.Context) {
	userID, _ := c.Get("user_id")

	userRoles, err := rbacService.GetUserRoles(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": userRoles})
}

func (s *Server) RevokeUserRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role id required"})
		return
	}

	err := rbacService.RevokeRole(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role revoked"})
}

// ============= JIT REQUESTS =============

func (s *Server) CreateJITRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input struct {
		Role            roles.Role `json:"role" binding:"required"`
		ResourceID      *string    `json:"resource_id"`
		DurationMinutes int        `json:"duration_minutes" binding:"required,min=1,max=1440"`
		Reason          string     `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request, err := jitService.CreateRequest(userID.(string), input.Role, input.ResourceID, input.DurationMinutes, input.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "JIT request created",
		"data":    request,
	})
}

func (s *Server) GetJITRequests(c *gin.Context) {
	userID := c.Query("user_id")

	var requests []roles.JITRequestDB
	var err error

	if userID != "" {
		requests, err = jitService.GetUserRequests(userID)
	} else {
		requests, err = jitService.GetPendingRequests()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func (s *Server) GetMyJITRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")

	requests, err := jitService.GetUserRequests(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func (s *Server) ApproveJITRequest(c *gin.Context) {
	requestID := c.Param("id")
	approverID, _ := c.Get("user_id")

	err := jitService.ApproveRequest(requestID, approverID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "request approved and role assigned"})
}

func (s *Server) RejectJITRequest(c *gin.Context) {
	requestID := c.Param("id")
	approverID, _ := c.Get("user_id")

	err := jitService.RejectRequest(requestID, approverID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "request rejected"})
}
