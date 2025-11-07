package dto

type RegisterDto struct {
	FullName          string `json:"full_name" binding:"required"`
	Username          string `json:"username" binding:"required"`
	Email             string `json:"email" binding:"required"`
	Password          string `json:"password" binding:"required,min=8"`
	ConfirmedPassword string `json:"confirmed_password" binding:"required"`
}
