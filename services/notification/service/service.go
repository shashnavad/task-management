// services/notification/service/service.go
//
// Changes from the original:
//   - NotificationService now holds a *WebSocketHub (Redis-backed) instead
//     of the old in-memory WebSocketHub.
//   - SendNotification calls hub.Publish (→ Redis → all pods) instead of
//     hub.Broadcast (→ only the local pod's connection map).
//   - NewNotificationService now requires a *WebSocketHub so callers (main.go)
//     wire up the Redis client once and pass the hub in.
//   - Everything else (in-memory notification store, HandleEvent, etc.)
//     is unchanged — those are orthogonal to the scaling concern.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/task-management/services/notification/models"
	"github.com/task-management/shared/events"
)

// NotificationService stores notifications in-memory and fans out real-time
// delivery via the Redis-backed WebSocketHub.
type NotificationService struct {
	mu            sync.Mutex
	notifications map[int]*models.Notification
	nextID        int

	// WsHub is exported so handlers can register / unregister connections.
	WsHub *WebSocketHub
}

// NewNotificationService wires the service to the provided hub.
// The hub must already be initialised (and its Run goroutine started)
// before the service is used.
func NewNotificationService(hub *WebSocketHub) *NotificationService {
	return &NotificationService{
		notifications: make(map[int]*models.Notification),
		nextID:        1,
		WsHub:         hub,
	}
}

// SendNotification persists the notification and publishes it through
// Redis so every pod can deliver it to the right WebSocket connection.
func (s *NotificationService) SendNotification(notification *models.Notification) (int, error) {
	s.mu.Lock()
	notification.ID = s.nextID
	s.nextID++
	notification.CreatedAt = time.Now()
	s.notifications[notification.ID] = notification
	s.mu.Unlock()

	// Publish via Redis so all pods fan-out to the correct connection.
	// We use a background context here; add a deadline if preferred.
	if err := s.WsHub.Publish(context.Background(), notification.UserID, notification); err != nil {
		log.Printf("hub: publish failed for user %d: %v", notification.UserID, err)
		// Non-fatal — the notification is already persisted.
	}

	return notification.ID, nil
}

func (s *NotificationService) GetNotifications(userID int) ([]*models.Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []*models.Notification
	for _, n := range s.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (s *NotificationService) MarkAsRead(notificationID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.notifications[notificationID]
	if !ok {
		return errors.New("notification not found")
	}
	n.IsRead = true
	return nil
}

func (s *NotificationService) DeleteNotification(notificationID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.notifications[notificationID]; !ok {
		return errors.New("notification not found")
	}
	delete(s.notifications, notificationID)
	return nil
}

func (s *NotificationService) HandleEvent(topic string, value []byte) error {
	switch topic {
	case "task.created":
		var event events.TaskCreatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			log.Printf("notification: unmarshal task.created: %v", err)
			return err
		}
		if event.AssigneeID != nil {
			n := &models.Notification{
				UserID:  *event.AssigneeID,
				Message: "You have been assigned a new task: " + event.Title,
				Type:    "task_assigned",
			}
			if _, err := s.SendNotification(n); err != nil {
				log.Printf("notification: send failed: %v", err)
			}
		}
	case "task.updated":
		var event events.TaskUpdatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			log.Printf("notification: unmarshal task.updated: %v", err)
			return err
		}
		if event.UserID != event.UpdatedBy {
			n := &models.Notification{
				UserID:  event.UpdatedBy,
				Message: "Task updated",
				Type:    "task_updated",
			}
			if _, err := s.SendNotification(n); err != nil {
				log.Printf("notification: send failed: %v", err)
			}
		}
	}
	return nil
}
