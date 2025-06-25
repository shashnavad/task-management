package service

import (
	"github.com/task-management/services/task/models"
	"github.com/task-management/services/task/repository"
	"github.com/task-management/shared/events"
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

func (s *TaskService) CreateTask(task *models.Task) error {
	err := s.repo.CreateTask(task)
	if err == nil {
		// Optionally emit event
		_ = s.eventProducer.Produce("task_created", task)
	}
	return err
}

func (s *TaskService) UpdateTask(task *models.Task) error {
	err := s.repo.UpdateTask(task)
	if err == nil {
		_ = s.eventProducer.Produce("task_updated", task)
	}
	return err
}

func (s *TaskService) DeleteTask(id int) error {
	err := s.repo.DeleteTask(id)
	if err == nil {
		_ = s.eventProducer.Produce("task_deleted", id)
	}
	return err
}

func (s *TaskService) AssignTask(id int, assigneeID int) error {
	err := s.repo.AssignTask(id, assigneeID)
	if err == nil {
		_ = s.eventProducer.Produce("task_assigned", map[string]interface{}{"task_id": id, "assignee_id": assigneeID})
	}
	return err
}

func (s *TaskService) UpdateStatus(id int, status string) error {
	err := s.repo.UpdateStatus(id, status)
	if err == nil {
		_ = s.eventProducer.Produce("task_status_updated", map[string]interface{}{"task_id": id, "status": status})
	}
	return err
}

func (s *TaskService) AddComment(comment *models.TaskComment) error {
	err := s.repo.AddComment(comment)
	if err == nil {
		_ = s.eventProducer.Produce("task_comment_added", comment)
	}
	return err
} 