// shared/events/events.go
package events

import (
	"fmt"
)

type TaskCreatedEvent struct {
	TaskID     int    `json:"task_id"`
	ProjectID  int    `json:"project_id"`
	AssigneeID *int   `json:"assignee_id"`
	Title      string `json:"title"`
	CreatedBy  int    `json:"created_by"`
	Timestamp  int64  `json:"timestamp"`
}

type TaskUpdatedEvent struct {
	TaskID    int    `json:"task_id"`
	Status    string `json:"status"`
	UpdatedBy int    `json:"updated_by"`
	Timestamp int64  `json:"timestamp"`
}

type ProjectCreatedEvent struct {
	ProjectID int    `json:"project_id"`
	Name      string `json:"name"`
	OwnerID   int    `json:"owner_id"`
	Timestamp int64  `json:"timestamp"`
}

// Producer defines an interface for producing events to a message broker or event bus.
type Producer interface {
	Produce(topic string, value interface{}) error
}

// mockProducer is a simple implementation of Producer that prints events to stdout.
type mockProducer struct{}

func (m *mockProducer) Produce(topic string, value interface{}) error {
	fmt.Printf("Produced event to topic %s: %+v\n", topic, value)
	return nil
}

// NewProducer returns a new mockProducer instance.
func NewProducer(brokers []string) (Producer, error) {
	return &mockProducer{}, nil
}
