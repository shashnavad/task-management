package saga

import (
	"context"
	"log"

	"github.com/task-management/shared/events"
)

// Saga represents a distributed transaction
type Saga struct {
	ID       string
	Steps    []SagaStep
	producer events.Producer
}

// SagaStep represents a step in the saga
type SagaStep struct {
	Name         string
	Action       func(ctx context.Context) error
	Compensation func(ctx context.Context) error
	Completed    bool
}

// NewSaga creates a new saga
func NewSaga(id string, producer events.Producer) *Saga {
	return &Saga{
		ID:       id,
		Steps:    []SagaStep{},
		producer: producer,
	}
}

// AddStep adds a step to the saga
func (s *Saga) AddStep(name string, action, compensation func(ctx context.Context) error) {
	s.Steps = append(s.Steps, SagaStep{
		Name:         name,
		Action:       action,
		Compensation: compensation,
	})
}

// Execute runs the saga
func (s *Saga) Execute(ctx context.Context) error {
	for i, step := range s.Steps {
		if err := step.Action(ctx); err != nil {
			log.Printf("Saga %s failed at step %s: %v", s.ID, step.Name, err)
			s.compensateFrom(ctx, i)
			return err
		}
		s.Steps[i].Completed = true
	}

	log.Printf("Saga %s completed successfully", s.ID)
	return nil
}

// compensateFrom rolls back from a failed step
func (s *Saga) compensateFrom(ctx context.Context, failedIndex int) {
	for i := failedIndex; i >= 0; i-- {
		step := &s.Steps[i]
		if step.Completed {
			if err := step.Compensation(ctx); err != nil {
				log.Printf("Compensation failed for step %s: %v", step.Name, err)
			}
		}
	}
}
