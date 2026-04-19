package service

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/task-management/services/notification/models"
	"github.com/task-management/shared/events"
)

type NotificationService struct {
	mu            sync.Mutex
	notifications map[int]*models.Notification
	nextID        int
	WsHub         *WebSocketHub
}

type WebSocketHub struct {
	clients map[int]map[*websocket.Conn]bool // userID -> set of connections
	mu      sync.Mutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[int]map[*websocket.Conn]bool),
	}
}

func (h *WebSocketHub) Register(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]bool)
	}
	h.clients[userID][conn] = true
}

func (h *WebSocketHub) Unregister(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] != nil {
		delete(h.clients[userID], conn)
		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *WebSocketHub) Broadcast(userID int, msg *models.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients[userID] {
		_ = conn.WriteJSON(msg)
	}
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		notifications: make(map[int]*models.Notification),
		nextID:        1,
		WsHub:         NewWebSocketHub(),
	}
}

func (s *NotificationService) SendNotification(notification *models.Notification) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	notification.ID = s.nextID
	s.nextID++
	notification.CreatedAt = time.Now()
	s.notifications[notification.ID] = notification
	// Broadcast to user in real time
	s.WsHub.Broadcast(notification.UserID, notification)
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
			log.Printf("Failed to unmarshal task.created event: %v", err)
			return err
		}
		// Notify assignee if exists
		if event.AssigneeID != nil {
			notification := &models.Notification{
				UserID:  *event.AssigneeID,
				Message: "You have been assigned a new task: " + event.Title,
				Type:    "task_assigned",
			}
			_, err := s.SendNotification(notification)
			if err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}
	case "task.updated":
		var event events.TaskUpdatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			log.Printf("Failed to unmarshal task.updated event: %v", err)
			return err
		}
		// Notify assignee or creator
		// For simplicity, notify user who initiated if different
		if event.UserID != event.UpdatedBy {
			notification := &models.Notification{
				UserID:  event.UpdatedBy,
				Message: "Task updated",
				Type:    "task_updated",
			}
			_, err := s.SendNotification(notification)
			if err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}
		// Add more cases as needed
	}
	return nil
}
