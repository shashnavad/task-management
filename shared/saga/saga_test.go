package saga

import (
	"context"
	"testing"
)

// MockProducer implements events.Producer for testing
type MockProducer struct {
	events []interface{}
}

func NewMockProducer() *MockProducer {
	return &MockProducer{
		events: []interface{}{},
	}
}

func (m *MockProducer) Produce(topic string, value interface{}) error {
	m.events = append(m.events, value)
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

// Tests
func TestSagaExecuteSuccess(t *testing.T) {
	producer := NewMockProducer()
	saga := NewSaga("test-saga", producer)

	stepExecuted := false
	saga.AddStep("test_step", func(ctx context.Context) error {
		stepExecuted = true
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	err := saga.Execute(context.Background())
	if err != nil {
		t.Errorf("Saga execution failed: %v", err)
	}

	if !stepExecuted {
		t.Errorf("Step was not executed")
	}
}

func TestSagaCompensation(t *testing.T) {
	producer := NewMockProducer()
	saga := NewSaga("test-saga", producer)

	compensationExecuted := false

	saga.AddStep("step1", func(ctx context.Context) error {
		return nil
	}, func(ctx context.Context) error {
		compensationExecuted = true
		return nil
	})

	saga.AddStep("step2", func(ctx context.Context) error {
		return context.Canceled // Simulate error
	}, func(ctx context.Context) error {
		return nil
	})

	err := saga.Execute(context.Background())
	if err == nil {
		t.Errorf("Expected error but saga succeeded")
	}

	if !compensationExecuted {
		t.Errorf("Compensation was not executed")
	}
}

func TestSagaMultipleSteps(t *testing.T) {
	producer := NewMockProducer()
	saga := NewSaga("test-saga", producer)

	executionOrder := []string{}

	saga.AddStep("step1", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "step1")
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	saga.AddStep("step2", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "step2")
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	saga.AddStep("step3", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "step3")
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	err := saga.Execute(context.Background())
	if err != nil {
		t.Errorf("Saga execution failed: %v", err)
	}

	if len(executionOrder) != 3 {
		t.Errorf("Not all steps were executed: %v", executionOrder)
	}
}
