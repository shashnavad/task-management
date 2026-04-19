// services/auth/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/auth/handlers"
	"github.com/task-management/services/auth/repository"
	"github.com/task-management/services/auth/service"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize database connection
	db := repository.InitDB()
	defer db.Close()

	// Initialize repository and service layers
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup routes
	router := gin.Default()

	auth := router.Group("/auth")
	{
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/signin", authHandler.SignIn)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	api := router.Group("/api/users")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/profile", authHandler.GetProfile)
		api.PUT("/profile", authHandler.UpdateProfile)
	}

	logger.Info("Auth service starting on port 8001")
	if err := http.ListenAndServe(":8001", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
