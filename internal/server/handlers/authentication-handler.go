package handlers

import (
	"AuthServer/internal/domain/dto"
	"AuthServer/internal/domain/models"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var verificationCodes = make(map[string]VerificationData)

type VerificationData struct {
	Code      string
	UserID    string
	ExpiresAt time.Time
}

func generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func getGmailService() (*gmail.Service, error) {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %v", err)
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return srv, nil
}

func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		return nil, fmt.Errorf("token not found - run OAuth setup first: %v", err)
	}
	return config.Client(context.Background(), tok), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func sendVerificationEmail(toEmail, code, userName string) error {
	srv, err := getGmailService()
	if err != nil {
		return fmt.Errorf("failed to get Gmail service: %v", err)
	}

	fromEmail := os.Getenv("SENDER_EMAIL")
	subject := "Email Verification Code"
	body := fmt.Sprintf(`

		<!DOCTYPE html>
			<html>
			<head>
				<style>
					body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
					.container { max-width: 600px; margin: 0 auto; padding: 20px; }
					.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
					.content { background-color: #f9f9f9; padding: 30px; border-radius: 5px; margin-top: 20px; }
					.code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; padding: 20px; background-color: white; border-radius: 5px; letter-spacing: 5px; }
					.footer { text-align: center; margin-top: 20px; color: #777; font-size: 12px; }
					.button { display: inline-block; padding: 12px 30px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
				</style>
			</head>
			<body>
				<div class="container">
					<div class="header">
						<h1>Email Verification</h1>
					</div>
					<div class="content">
						<h2>Hello %s!</h2>
						<p>Thank you for registering. Please verify your email address using the code below:</p>
						<div class="code">%s</div>
						<p style="text-align: center; margin-top: 20px;">Or click the button below to verify:</p>
						<div style="text-align: center;">
							<a href="http://localhost:3000/verify?code=%s" class="button">Verify Email</a>
						</div>
						<p style="margin-top: 30px; font-size: 14px; color: #666;">
							This code will expire in 15 minutes. If you didn't register for this account, please ignore this email.
						</p>
					</div>
					<div class="footer">
						<p>This is an automated message, please do not reply.</p>
					</div>
				</div>
			</body>
			</html>
		`, userName, code, code)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s",
		fromEmail, toEmail, subject, body)

	encoded := base64.URLEncoding.EncodeToString([]byte(message))
	gmailMessage := &gmail.Message{Raw: encoded}

	_, err = srv.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *Server) Register(c *gin.Context) {
	var registerData dto.RegisterDto

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

	registerData.FullName = string(decodedFullName)
	registerData.Username = string(decodedUsername)
	registerData.Email = string(decodedEmail)
	registerData.Password = string(decodedPassword)
	registerData.ConfirmedPassword = string(decodedConfirmedPassword)

	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRegex.MatchString(registerData.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	log.Println("Decoded data - Username:", registerData.Username, "Email:", registerData.Email)

	if registerData.Password != registerData.ConfirmedPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	hashedPassword, err := hashService.HashPassword(registerData.Password)
	if err != nil {
		log.Println("Password hashing error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

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
		log.Printf("Error saving user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	verificationCode := generateVerificationCode()

	verificationCodes[user.Email] = VerificationData{
		Code:      verificationCode,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = sendVerificationEmail(user.Email, verificationCode, user.FullName)
	if err != nil {
		log.Printf("Error sending verification email: %v", err)
		log.Println("Warning: User registered but verification email failed to send")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":               fmt.Sprintf("Registration successful %s. Please check your email for verification code.", user.FullName),
		"user_id":               user.ID,
		"email":                 user.Email,
		"requires_verification": true,
	})
}

func (s *Server) VerificationCode(c *gin.Context) {
	var verificationPacket struct {
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&verificationPacket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	code := verificationPacket.VerificationCode

	if len(code) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code must be 6 digits"})
		return
	}

	var foundEmail string
	var foundData VerificationData
	var found bool

	for email, data := range verificationCodes {
		if data.Code == code {
			foundEmail = email
			foundData = data
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	if time.Now().After(foundData.ExpiresAt) {
		delete(verificationCodes, foundEmail)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code has expired"})
		return
	}

	delete(verificationCodes, foundEmail)

	accessToken := tokenService.GenerateAccessToken(foundData.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Email verified successfully",
		"access_token": accessToken,
	})
}

func (s *Server) VerifyEmail(c *gin.Context) {
	var verifyData struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&verifyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	verificationData, exists := verificationCodes[verifyData.Email]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No verification code found for this email"})
		return
	}

	if time.Now().After(verificationData.ExpiresAt) {
		delete(verificationCodes, verifyData.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code has expired"})
		return
	}

	if verificationData.Code != verifyData.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	delete(verificationCodes, verifyData.Email)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Email verified successfully",
		"access_token": tokenService.GenerateAccessToken(verificationData.UserID),
	})
}

func (s *Server) ResendVerificationCode(c *gin.Context) {
	var resendData struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&resendData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := userRepo.FindByEmailOrUsername(resendData.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	verificationCode := generateVerificationCode()
	verificationCodes[user.Email] = VerificationData{
		Code:      verificationCode,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = sendVerificationEmail(user.Email, verificationCode, user.FullName)
	if err != nil {
		log.Printf("Error sending verification email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent successfully",
	})
}

func (s *Server) Login(c *gin.Context) {
	var loginData dto.LoginDto

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

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
		log.Printf("User not found: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong credentials"})
		return
	}

	storedPassword := existingUser.Password
	loginPassword := loginData.Password

	if hashService.VerifyPassword(loginPassword, storedPassword) {
		verificationCode := generateVerificationCode()

		verificationCodes[existingUser.Email] = VerificationData{
			Code:      verificationCode,
			UserID:    existingUser.ID,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		err = sendVerificationEmail(existingUser.Email, verificationCode, existingUser.FullName)
		if err != nil {
			log.Printf("Error sending verification email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":               fmt.Sprintf("Login verification sent to %s", existingUser.Email),
			"email":                 existingUser.Email,
			"requires_verification": true,
		})
		return
	} else {
		log.Println("Password does not match")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong credentials",
		})
		return
	}
}
