package models

import "time"

type User struct {
	ID        string    `json:"-"`
	FullName  string    `json:"full_name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

//RefreshToken           string        `gorm:"size:255"`
//RefreshTokenExpiryDate time.Time     // 30 days after each log in
