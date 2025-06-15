// services/auth/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/auth/handlers"
	"github.com/task-management/services/auth/repository"
	"github.com/task-management/services/auth/service"
)

func main() {
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
	api.Use(authHandler.AuthMiddleware())
	{
		api.GET("/profile", authHandler.GetProfile)
		api.PUT("/profile", authHandler.UpdateProfile)
	}

	log.Println("Auth service starting on port 8001")
	log.Fatal(http.ListenAndServe(":8001", router))
}
