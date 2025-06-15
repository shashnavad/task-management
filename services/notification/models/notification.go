// services/notification/models/notification.go
package models

import "time"

type Notification struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Data      string    `json:"data" db:"data"` // JSON string for additional data
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	UserID  int         `json:"user_id"`
	Title   string      `json:"title"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
