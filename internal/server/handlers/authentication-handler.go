package handlers

import (
	"AuthServer/internal/domain/dto"
	"AuthServer/internal/domain/models"
	"encoding/base64"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"fmt"
)

func (s *Server) Register(c *gin.Context) {
	var registerData dto.RegisterDto

	// Bind JSON request body
	if err := c.ShouldBindJSON(&registerData); err != nil {
		log.Println("JSON binding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Println("Received encoded data:", registerData)

	// Decode base64 fields
	decodedFullName, err := base64.StdEncoding.DecodeString(registerData.FullName)
	if err != nil {
		log.Println("Full name decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid full_name encoding"})
		return
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(registerData.Username)
	if err != nil {
		log.Println("Username decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username encoding"})
		return
	}

	decodedEmail, err := base64.StdEncoding.DecodeString(registerData.Email)
	if err != nil {
		log.Println("Email decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email encoding"})
		return
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(registerData.Password)
	if err != nil {
		log.Println("Password decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password encoding"})
		return
	}

	decodedConfirmedPassword, err := base64.StdEncoding.DecodeString(registerData.ConfirmedPassword)
	if err != nil {
		log.Println("Confirmed password decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid confirmed_password encoding"})
		return
	}

	// Convert decoded bytes to strings
	registerData.FullName = string(decodedFullName)
	registerData.Username = string(decodedUsername)

	registerData.Email = string(decodedEmail)
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRegex.MatchString(registerData.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	registerData.Password = string(decodedPassword)
	registerData.ConfirmedPassword = string(decodedConfirmedPassword)

	log.Println("Decoded data - Username:", registerData.Username, "Email:", registerData.Email)

	if registerData.Password != registerData.ConfirmedPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	hashedPassword, err := hashService.HashPassword(registerData.Password)

	var user = models.User{
		ID:        uuid.New().String(),
		FullName:  registerData.FullName,
		Username:  registerData.Username,
		Email:     registerData.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	err = userRepo.Save(user)

	if err != nil {
		log.Fatalf("Error saving user: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Registration successful %s", user.FullName),
		"access_token": tokenService.GenerateAccessToken(user.ID),
	})
	return
}

func (s *Server) Login(c *gin.Context) {
	var loginData dto.LoginDto

	if err := c.ShouldBindJSON(&loginData); err != nil {
		//log.Fatalf("JSON binding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	//log.Println("Received encoded data:", loginData)

	identifier, err := base64.StdEncoding.DecodeString(loginData.Identifier)
	if err != nil {
		log.Println("Username or email decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username_or_email encoding"})
		return
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(loginData.Password)
	if err != nil {
		log.Println("Password decoding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password encoding"})
		return
	}

	loginData.Identifier = string(identifier)
	loginData.Password = string(decodedPassword)

	existingUser, err := userRepo.FindByEmailOrUsername(loginData.Identifier)
	if err != nil {
		log.Fatalf(err.Error())
	}

	storedPassword := existingUser.Password
	loginPassword := loginData.Password

	if hashService.VerifyPassword(loginPassword, storedPassword) {
		c.JSON(http.StatusOK, gin.H{
			"message":      fmt.Sprintf("Login successful %s", existingUser.FullName),
			"access_token": tokenService.GenerateAccessToken(existingUser.ID),
		})
		return
	} else {
		log.Println("Password does not match")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong credentials",
		})
		return
	}

}
