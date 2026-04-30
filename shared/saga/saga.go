package saga
package saga

import (
	"context"
	"fmt"
	"log"

	"github.com/task-management/shared/events"








































































}	}		}			span.End()			}				log.Printf("Compensation failed for step %s: %v", step.Name, err)			if err := step.Compensation(stepCtx); err != nil {			stepCtx, span := s.tracer.Start(ctx, fmt.Sprintf("saga.compensate.%s", step.Name))		if step.Completed {		step := &s.Steps[i]	for i := failedIndex; i >= 0; i-- {func (s *Saga) compensateFrom(ctx context.Context, failedIndex int) {// compensateFrom rolls back from a failed step}	return nil	log.Printf("Saga %s completed successfully", s.ID)	}		s.Steps[i].Completed = true		stepSpan.End()		}			return err			s.compensateFrom(stepCtx, i)			log.Printf("Saga %s failed at step %s: %v", s.ID, step.Name, err)			stepSpan.End()		if err := step.Action(stepCtx); err != nil {		stepCtx, stepSpan := s.tracer.Start(ctx, fmt.Sprintf("saga.step.%s", step.Name))	for i, step := range s.Steps {	defer span.End()	ctx, span := s.tracer.Start(ctx, "saga.execute")func (s *Saga) Execute(ctx context.Context) error {// Execute runs the saga}	})		Compensation: compensation,		Action:       action,		Name:         name,	s.Steps = append(s.Steps, SagaStep{func (s *Saga) AddStep(name string, action, compensation func(ctx context.Context) error) {// AddStep adds a step to the saga}	}		tracer:   tracer,		producer: producer,		Steps:    []SagaStep{},		ID:       id,	return &Saga{func NewSaga(id string, producer events.Producer, tracer trace.Tracer) *Saga {// NewSaga creates a new saga}	Completed    bool	Compensation func(ctx context.Context) error	Action       func(ctx context.Context) error	Name         stringtype SagaStep struct {// SagaStep represents a step in the saga}	tracer   trace.Tracer	producer events.Producer	Steps    []SagaStep	ID       stringtype Saga struct {// Saga represents a distributed transaction)	"go.opentelemetry.io/otel/trace"	pb "github.com/task-management/proto/notification"