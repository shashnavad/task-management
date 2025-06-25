// service/auth_service.go

package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/task-management/services/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// AuthServiceInterface defines the methods for authentication service
type AuthServiceInterface interface {
	CreateUser(email, username, password, fullName string) (string, error)
	AuthenticateUser(username, password string) (string, error)
	RefreshAccessToken(refreshToken string) (string, error)
	BlacklistToken(token string) error
	ValidateToken(token string) (string, error)
	IsTokenBlacklisted(token string) (bool, error)
	GetUserByID(userID string) (*User, error)
	UpdateUserProfile(userID, fullName, email string) error
}

// AuthService implements AuthServiceInterface
type AuthService struct {
	userRepo       repository.UserRepositoryInterface
	jwtSecret      string
	tokenBlacklist map[string]time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(repo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:       repo,
		jwtSecret:      "your-secret-key", // In a real app, this should be loaded from env/config
		tokenBlacklist: make(map[string]time.Time),
	}
}

// CreateUser creates a new user
func (s *AuthService) CreateUser(email, username, password, fullName string) (string, error) {
	// Check if user already exists
	exists, err := s.userRepo.CheckUserExists(username, email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Create user
	userID, err := s.userRepo.CreateUser(email, username, string(hashedPassword), fullName)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// AuthenticateUser authenticates a user and returns a JWT token
func (s *AuthService) AuthenticateUser(username, password string) (string, error) {
	// Get user from database
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours

	// Sign the token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RefreshAccessToken validates a refresh token and issues a new access token
func (s *AuthService) RefreshAccessToken(refreshToken string) (string, error) {
	// Validate refresh token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	// Check if token is in blacklist
	isBlacklisted, err := s.IsTokenBlacklisted(refreshToken)
	if err != nil {
		return "", err
	}
	if isBlacklisted {
		return "", errors.New("token has been revoked")
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Get user from database
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return "", err
	}

	// Generate new access token
	newToken := jwt.New(jwt.SigningMethodHS256)
	newClaims := newToken.Claims.(jwt.MapClaims)
	newClaims["user_id"] = user.ID
	newClaims["username"] = user.Username
	newClaims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours

	// Sign the token
	tokenString, err := newToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// BlacklistToken adds a token to the blacklist
func (s *AuthService) BlacklistToken(token string) error {
	// Add token to blacklist with expiration time
	s.tokenBlacklist[token] = time.Now().Add(time.Hour * 24) // Keep in blacklist for 24 hours
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *AuthService) IsTokenBlacklisted(token string) (bool, error) {
	// Clean up expired tokens
	for t, exp := range s.tokenBlacklist {
		if time.Now().After(exp) {
			delete(s.tokenBlacklist, t)
		}
	}

	// Check if token is in blacklist
	_, exists := s.tokenBlacklist[token]
	return exists, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	// Parse token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		// Try to convert from float64 (if stored as int in DB)
		if idFloat, ok := claims["user_id"].(float64); ok {
			userID = fmt.Sprintf("%d", int(idFloat))
		} else {
			return "", errors.New("invalid token claims")
		}
	}

	return userID, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*User, error) {
	// Get user from database
	dbUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Map to DTO
	user := &User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		FullName: dbUser.FullName,
	}

	return user, nil
}

// UpdateUserProfile updates a user's profile
func (s *AuthService) UpdateUserProfile(userID, fullName, email string) error {
	// Update user in database
	return s.userRepo.UpdateUser(userID, fullName, email)
}
