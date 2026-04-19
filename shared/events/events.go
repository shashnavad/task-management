// shared/events/events.go
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type TaskCreatedEvent struct {
	TaskID     int    `json:"task_id"`
	ProjectID  int    `json:"project_id"`
	AssigneeID *int   `json:"assignee_id"`
	Title      string `json:"title"`
	CreatedBy  int    `json:"created_by"`
	UserID     int    `json:"user_id"` // Added for JWT context
	Timestamp  int64  `json:"timestamp"`
}

type TaskUpdatedEvent struct {
	TaskID    int    `json:"task_id"`
	Status    string `json:"status"`
	UpdatedBy int    `json:"updated_by"`
	UserID    int    `json:"user_id"` // Added for JWT context
	Timestamp int64  `json:"timestamp"`
}

type ProjectCreatedEvent struct {
	ProjectID int    `json:"project_id"`
	Name      string `json:"name"`
	OwnerID   int    `json:"owner_id"`
	UserID    int    `json:"user_id"` // Added for JWT context
	Timestamp int64  `json:"timestamp"`
}

// Producer defines an interface for producing events to a message broker or event bus.
type Producer interface {
	Produce(topic string, value interface{}) error
	Close() error
}

// Consumer defines an interface for consuming events.
type Consumer interface {
	ConsumeEvents(topics []string, handler func(topic string, value []byte) error) error
	Close() error
}

// kafkaProducer implements Producer using Kafka.
type kafkaProducer struct {
	writer *kafka.Writer
}

func (k *kafkaProducer) Produce(topic string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Value: data,
	})
}

func (k *kafkaProducer) Close() error {
	return k.writer.Close()
}

// kafkaConsumer implements Consumer using Kafka.
type kafkaConsumer struct {
	reader *kafka.Reader
}

func (k *kafkaConsumer) ConsumeEvents(topics []string, handler func(topic string, value []byte) error) error {
	// For simplicity, consume from the first topic; in production, handle multiple.
	if len(topics) == 0 {
		return fmt.Errorf("no topics provided")
	}
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "notification-group",
		Topic:   topics[0], // Consume from first topic
	})
	defer k.reader.Close()

	for {
		msg, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}
		if err := handler(msg.Topic, msg.Value); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

func (k *kafkaConsumer) Close() error {
	if k.reader != nil {
		return k.reader.Close()
	}
	return nil
}

// NewProducer returns a new Kafka producer instance.
func NewProducer(brokers []string) (Producer, error) {
	return &kafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}, nil
}

// NewConsumer returns a new Kafka consumer instance.
func NewConsumer(brokers []string, groupID string) (Consumer, error) {
	return &kafkaConsumer{}, nil
}
