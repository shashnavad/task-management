package service

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/task-management/services/notification/models"
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
