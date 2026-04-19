// services/notification/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/notification/handlers"
	"github.com/task-management/services/notification/service"
	"github.com/task-management/shared/events"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize notification service
	notificationService := service.NewNotificationService()
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Initialize event consumer
	consumer, err := events.NewConsumer([]string{"localhost:9092"}, "notification-group")
	if err != nil {
		logger.Fatal("Failed to create event consumer", zap.Error(err))
	}
	go consumer.ConsumeEvents([]string{"task.created", "task.updated", "project.created"}, notificationService.HandleEvent)

	// Setup routes
	router := gin.Default()

	// WebSocket endpoint
	router.GET("/ws", notificationHandler.HandleWebSocket)

	api := router.Group("/api/notifications")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("", notificationHandler.GetNotifications)
		api.PUT("/:id/read", notificationHandler.MarkAsRead)
		api.POST("/send", notificationHandler.SendNotification)
	}

	logger.Info("Notification service starting on port 8005")
	if err := http.ListenAndServe(":8005", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
