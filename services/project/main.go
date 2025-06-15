// services/project/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/project/handlers"
	"github.com/task-management/project/repository"
	"github.com/task-management/project/service"
)

func main() {
	// Initialize database connection
	db := repository.InitDB()
	defer db.Close()

	// Initialize layers
	projectRepo := repository.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	// Setup routes
	router := gin.Default()

	api := router.Group("/api/projects")
	api.Use(authMiddleware()) // JWT validation middleware
	{
		api.GET("", projectHandler.GetProjects)
		api.POST("", projectHandler.CreateProject)
		api.GET("/:id", projectHandler.GetProject)
		api.PUT("/:id", projectHandler.UpdateProject)
		api.DELETE("/:id", projectHandler.DeleteProject)
		api.POST("/:id/members", projectHandler.AddMember)
		api.DELETE("/:id/members/:userId", projectHandler.RemoveMember)
	}

	log.Println("Project service starting on port 8002")
	log.Fatal(http.ListenAndServe(":8002", router))
}

func authMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// JWT token validation logic here
		c.Next()
	})
}
