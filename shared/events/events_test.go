package events

import (
	"testing"
)

// MockProducer implements Producer for testing
type MockProducer struct {
	messages []map[string]interface{}
}

func NewMockProducer() *MockProducer {
	return &MockProducer{
		messages: []map[string]interface{}{},
	}
}

func (m *MockProducer) Produce(topic string, value interface{}) error {
	msg := map[string]interface{}{
		"topic": topic,
		"value": value,
	}
	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

// Tests
func TestEventProducer(t *testing.T) {
	producer := NewMockProducer()

	event := TaskCreatedEvent{
		TaskID:    1,
		ProjectID: 1,
		Title:     "Test Task",
		CreatedBy: 1,
		UserID:    1,
	}

	err := producer.Produce("task.created", event)
	if err != nil {
		t.Errorf("Failed to produce event: %v", err)
	}

	if len(producer.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(producer.messages))
	}

	if producer.messages[0]["topic"] != "task.created" {
		t.Errorf("Topic mismatch: got %v, want task.created", producer.messages[0]["topic"])
	}
}

func TestMultipleEvents(t *testing.T) {
	producer := NewMockProducer()

	for i := 1; i <= 5; i++ {
		event := TaskCreatedEvent{
			TaskID:    i,
			ProjectID: 1,
			Title:     "Task " + string(byte(i)),
			CreatedBy: 1,
			UserID:    1,
		}
		err := producer.Produce("task.created", event)
		if err != nil {
			t.Errorf("Failed to produce event %d: %v", i, err)
		}
	}

	if len(producer.messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(producer.messages))
	}
}
