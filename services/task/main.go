// services/task/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/task/handlers"
	"github.com/task-management/services/task/repository"
	"github.com/task-management/services/task/service"
	"github.com/task-management/shared/events"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize database and event producer
	db := repository.InitDB()
	defer db.Close()

	eventProducer, err := events.NewProducer([]string{"localhost:9092"})
	if err != nil {
		logger.Fatal("Failed to create event producer", zap.Error(err))
	}
	defer eventProducer.Close()

	// Initialize layers
	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo, eventProducer)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Setup routes
	router := gin.Default()

	api := router.Group("/api/tasks")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("", taskHandler.GetTasks)
		api.POST("", taskHandler.CreateTask)
		api.GET("/:id", taskHandler.GetTask)
		api.PUT("/:id", taskHandler.UpdateTask)
		api.DELETE("/:id", taskHandler.DeleteTask)
		api.PUT("/:id/assign", taskHandler.AssignTask)
		api.PUT("/:id/status", taskHandler.UpdateStatus)
		api.POST("/:id/comments", taskHandler.AddComment)
	}

	logger.Info("Task service starting on port 8003")
	if err := http.ListenAndServe(":8003", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
