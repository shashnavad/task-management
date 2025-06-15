package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/task-management-system/services/auth/service"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	type SignUpRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		FullName string `json:"full_name" binding:"required"`
	}
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// Call service layer
	userID, err := h.service.CreateUser(req.Email, req.Username, req.Password, req.FullName)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"id": userID})
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	type SignInRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, err := h.service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(200, gin.H{"token": token})
}
