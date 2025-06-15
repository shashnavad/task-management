// services/auth/models/user.go
package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email" binding:"required,email"`
	Username     string    `json:"username" db:"username" binding:"required"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"full_name" db:"full_name" binding:"required"`
	Role         string    `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
}

func (u *User) HashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) ValidatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
