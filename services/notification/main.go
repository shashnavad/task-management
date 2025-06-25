// services/notification/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/notification/handlers"
	"github.com/task-management/services/notification/service"
	"github.com/task-management/shared/middleware"
)

func main() {
	// Initialize notification service
	notificationService := service.NewNotificationService()
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Optionally, initialize event consumer if needed
	// consumer, err := events.NewConsumer([]string{"localhost:9092"}, "notification-group")
	// if err != nil {
	// 	log.Fatal("Failed to create event consumer:", err)
	// }
	// go consumer.ConsumeEvents([]string{"task.created", "task.updated", "project.created"}, notificationService.HandleEvent)

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

	log.Println("Notification service starting on port 8005")
	log.Fatal(http.ListenAndServe(":8005", router))
}
