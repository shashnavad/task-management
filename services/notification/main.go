// services/notification/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/notification/handlers"
	"github.com/task-management/notification/service"
	"github.com/task-management/shared/events"
)

func main() {
	// Initialize WebSocket hub
	hub := service.NewHub()
	go hub.Run()

	// Initialize event consumer
	consumer, err := events.NewConsumer([]string{"localhost:9092"}, "notification-group")
	if err != nil {
		log.Fatal("Failed to create event consumer:", err)
	}

	// Initialize notification service
	notificationService := service.NewNotificationService(hub)
	notificationHandler := handlers.NewNotificationHandler(notificationService, hub)

	// Start consuming events
	go consumer.ConsumeEvents([]string{"task.created", "task.updated", "project.created"}, notificationService.HandleEvent)

	// Setup routes
	router := gin.Default()

	// WebSocket endpoint
	router.GET("/ws", notificationHandler.HandleWebSocket)

	api := router.Group("/api/notifications")
	api.Use(authMiddleware())
	{
		api.GET("", notificationHandler.GetNotifications)
		api.PUT("/:id/read", notificationHandler.MarkAsRead)
		api.POST("/send", notificationHandler.SendNotification)
	}

	log.Println("Notification service starting on port 8005")
	log.Fatal(http.ListenAndServe(":8005", router))
}
