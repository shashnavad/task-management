package service

import (
	"github.com/task-management-system/services/auth/models"
	"github.com/task-management-system/services/auth/repository"
)

type AuthService interface {
	CreateUser(email, username, password, fullName string) (int, error)
	AuthenticateUser(username, password string) (string, error)
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) CreateUser(email, username, password, fullName string) (int, error) {
	// Hash password, then call repo
	return s.repo.CreateUser(email, username, password, fullName) // NOTE: In practice, hash password first!
}

func (s *authService) AuthenticateUser(username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	// In practice, validate password hash
	if user.PasswordHash == "mock-hash" && password == "mock-password" { // Mock logic!
		return "mock-jwt-token", nil
	}
	return "", models.ErrInvalidCredentials
}
