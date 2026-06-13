// services/notification/main.go
//
// Diff from original:
//   - Reads REDIS_ADDR env var (defaults to localhost:6379).
//   - Constructs a *redis.Client and passes it to NewWebSocketHub.
//   - Starts hub.Run in a goroutine so the Redis subscriber is active
//     before any HTTP/WS traffic arrives.
//   - Passes the hub into NewNotificationService instead of constructing
//     an in-memory hub internally.

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/task-management/services/notification/handlers"
	"github.com/task-management/services/notification/service"
	"github.com/task-management/shared/events"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func redisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

func kafkaBrokers() []string {
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		return []string{brokers}
	}
	return []string{"localhost:9092"}
}

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// --- Redis ---
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr()})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}

	// --- WebSocket hub (Redis-backed) ---
	hub := service.NewWebSocketHub(rdb, logger)
	go hub.Run(context.Background()) // starts the Redis subscriber goroutine

	// --- Notification service ---
	notificationService := service.NewNotificationService(hub)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// --- Kafka consumer ---
	consumer, err := events.NewConsumer(kafkaBrokers(), "notification-group")
	if err != nil {
		logger.Fatal("failed to create event consumer", zap.Error(err))
	}
	go consumer.ConsumeEvents(
		[]string{"task.created", "task.updated", "project.created"},
		notificationService.HandleEvent,
	)

	// --- HTTP/WS routes ---
	router := gin.Default()
	router.GET("/ws", notificationHandler.HandleWebSocket)

	api := router.Group("/api/notifications")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("", notificationHandler.GetNotifications)
		api.PUT("/:id/read", notificationHandler.MarkAsRead)
		api.POST("/send", notificationHandler.SendNotification)
	}

	logger.Info("notification service starting on :8005")
	if err := http.ListenAndServe(":8005", router); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
