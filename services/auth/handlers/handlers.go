// handlers/auth_handlers.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/auth/service"
)

// AuthHandler struct holds the authentication service
type AuthHandler struct {
	service service.AuthServiceInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(s service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{service: s}
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(c *gin.Context) {
	type SignUpRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		FullName string `json:"full_name" binding:"required"`
	}

	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service layer
	userID, err := h.service.CreateUser(req.Email, req.Username, req.Password, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": userID})
}

// SignIn handles user authentication
func (h *AuthHandler) SignIn(c *gin.Context) {
	type SignInRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and refresh token
	newToken, err := h.service.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newToken})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization header required"})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Blacklist the token
	err := h.service.BlacklistToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

// GetProfile retrieves the user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	type UpdateProfileRequest struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateUserProfile(userID.(string), req.FullName, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

// AuthMiddleware provides JWT authentication for protected routes
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token
		userID, err := h.service.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := h.service.IsTokenBlacklisted(tokenString)
		if err != nil || isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		// Set user ID in context for use in handlers
		c.Set("userID", userID)
		c.Next()
	}
}
