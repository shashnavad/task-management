package service

import (
	"context"
	"fmt"

	"github.com/task-management/services/task/models"
	"github.com/task-management/services/task/repository"
	"github.com/task-management/shared/events"
	"github.com/task-management/shared/saga"
)

type TaskService struct {
	repo          *repository.TaskRepository
	eventProducer events.Producer
}

func NewTaskService(repo *repository.TaskRepository, eventProducer events.Producer) *TaskService {
	return &TaskService{repo: repo, eventProducer: eventProducer}
}

func (s *TaskService) GetTasks() ([]*models.Task, error) {
	return s.repo.GetTasks()
}

func (s *TaskService) GetTask(id int) (*models.Task, error) {
	return s.repo.GetTask(id)
}

func (s *TaskService) CreateTask(task *models.Task, userID int) error {
	ctx := context.Background()
	sagaObj := saga.NewSaga(fmt.Sprintf("task-creation-%d", task.ID), s.eventProducer)

	// Step 1: Create task in database
	sagaObj.AddStep("create_task", func(ctx context.Context) error {
		return s.repo.CreateTask(task)
	}, func(ctx context.Context) error {
		return s.repo.DeleteTask(task.ID) // Compensation: delete task
	})

	// Step 2: Publish event for reporting
	sagaObj.AddStep("publish_event", func(ctx context.Context) error {
		event := events.TaskCreatedEvent{
			TaskID:     task.ID,
			ProjectID:  task.ProjectID,
			AssigneeID: task.AssigneeID,
			Title:      task.Title,
			CreatedBy:  task.CreatedBy,
			UserID:     userID,
			Timestamp:  task.CreatedAt.Unix(),
		}
		return s.eventProducer.Produce("task.created", event)
	}, func(ctx context.Context) error {
		// Compensation: publish a task.deleted event
		return s.eventProducer.Produce("task.deleted", map[string]interface{}{"task_id": task.ID})
	})

	return sagaObj.Execute(ctx)
}

func (s *TaskService) UpdateTask(task *models.Task, userID int) error {
	err := s.repo.UpdateTask(task)
	if err == nil {
		event := events.TaskUpdatedEvent{
			TaskID:    task.ID,
			Status:    task.Status,
			UpdatedBy: task.UpdatedBy,
			UserID:    userID,
			Timestamp: task.UpdatedAt.Unix(),
		}
		_ = s.eventProducer.Produce("task.updated", event)
	}
	return err
}

func (s *TaskService) DeleteTask(id int, userID int) error {
	err := s.repo.DeleteTask(id)
	if err == nil {
		_ = s.eventProducer.Produce("task.deleted", map[string]interface{}{"task_id": id, "user_id": userID})
	}
	return err
}

func (s *TaskService) AssignTask(id int, assigneeID int, userID int) error {
	err := s.repo.AssignTask(id, assigneeID)
	if err == nil {
		_ = s.eventProducer.Produce("task.assigned", map[string]interface{}{"task_id": id, "assignee_id": assigneeID, "user_id": userID})
	}
	return err
}

func (s *TaskService) UpdateStatus(id int, status string, userID int) error {
	err := s.repo.UpdateStatus(id, status)
	if err == nil {
		_ = s.eventProducer.Produce("task.status_updated", map[string]interface{}{"task_id": id, "status": status, "user_id": userID})
	}
	return err
}

func (s *TaskService) AddComment(comment *models.TaskComment, userID int) error {
	err := s.repo.AddComment(comment)
	if err == nil {
		_ = s.eventProducer.Produce("task.comment_added", map[string]interface{}{"comment": comment, "user_id": userID})
	}
	return err
}
