package service

import (
	"github.com/task-management/services/project/models"
	"github.com/task-management/services/project/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(p *models.Project) (int, error) {
	return s.repo.CreateProject(p)
}

func (s *ProjectService) GetProject(id int) (*models.Project, error) {
	return s.repo.GetProject(id)
}

func (s *ProjectService) UpdateProject(id int, update *models.Project) error {
	return s.repo.UpdateProject(id, update)
}

func (s *ProjectService) DeleteProject(id int) error {
	return s.repo.DeleteProject(id)
}

func (s *ProjectService) ListProjects() []*models.Project {
	return s.repo.ListProjects()
}

func (s *ProjectService) AddMember(projectID int, member *models.ProjectMember) error {
	return s.repo.AddMember(projectID, member)
}

func (s *ProjectService) RemoveMember(projectID, userID int) error {
	return s.repo.RemoveMember(projectID, userID)
}

func (s *ProjectService) ListMembers(projectID int) ([]*models.ProjectMember, error) {
	return s.repo.ListMembers(projectID)
}
