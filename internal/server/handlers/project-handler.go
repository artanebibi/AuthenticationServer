package handlers

import (
	"AuthServer/internal/domain/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetAllProjects(c *gin.Context) {
	user, _ := getUserFromDatabase(c)
	if user == nil {
		return
	}

	projects, err := projectService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": projects,
	})
}

func (s *Server) GetProjectById(c *gin.Context) {
	user, _ := getUserFromDatabase(c)
	if user == nil {
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project id is required"})
		return
	}

	project, err := projectService.FindById(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": project,
	})
}

func (s *Server) CreateProject(c *gin.Context) {
	user, _ := getUserFromDatabase(c)
	if user == nil {
		return
	}

	var input struct {
		Id          string `json:"id"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := models.Project{
		ID:          input.Id,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now(),
	}

	savedProject, err := projectService.Save(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "project created successfully",
		"data":    savedProject,
	})
}

func (s *Server) UpdateProject(c *gin.Context) {
	user, _ := getUserFromDatabase(c)
	if user == nil {
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project id is required"})
		return
	}

	// Check if project exists
	existingProject, err := projectService.FindById(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	if input.Name != "" {
		existingProject.Name = input.Name
	}
	if input.Description != "" {
		existingProject.Description = input.Description
	}

	if err := projectService.Update(*existingProject); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project updated successfully",
		"data":    existingProject,
	})
}

func (s *Server) DeleteProject(c *gin.Context) {
	user, _ := getUserFromDatabase(c)
	if user == nil {
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project id is required"})
		return
	}

	if err := projectService.Delete(projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project deleted successfully",
	})
}
