// services/project/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/project/handlers"
	"github.com/task-management/services/project/repository"
	"github.com/task-management/services/project/service"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize in-memory repository, service, and handler
	repo := repository.NewProjectRepository()
	svc := service.NewProjectService(repo)
	handler := handlers.NewProjectHandler(svc)

	// Setup routes
	router := gin.Default()

	api := router.Group("/api/projects")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("", handler.GetProjects)
		api.POST("", handler.CreateProject)
		api.GET("/:id", handler.GetProject)
		api.PUT("/:id", handler.UpdateProject)
		api.DELETE("/:id", handler.DeleteProject)
		api.POST("/:id/members", handler.AddMember)
		api.DELETE("/:id/members/:userId", handler.RemoveMember)
	}

	logger.Info("Project service starting on port 8002")
	if err := http.ListenAndServe(":8002", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func InitProjectDB() *repository.ProjectRepository {
	return repository.NewProjectRepository()
}

func InitTaskDB() *repository.ProjectRepository {
	return repository.NewProjectRepository()
}
