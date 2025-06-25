package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/task-management/services/project/models"
)

type ProjectRepository struct {
	mu       sync.Mutex
	projects map[int]*models.Project
	members  map[int][]*models.ProjectMember // projectID -> members
	nextID   int
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[int]*models.Project),
		members:  make(map[int][]*models.ProjectMember),
		nextID:   1,
	}
}

func (r *ProjectRepository) CreateProject(p *models.Project) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.ID = r.nextID
	r.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	r.projects[p.ID] = p
	return p.ID, nil
}

func (r *ProjectRepository) GetProject(id int) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.projects[id]
	if !ok {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (r *ProjectRepository) UpdateProject(id int, update *models.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.projects[id]
	if !ok {
		return errors.New("project not found")
	}
	p.Name = update.Name
	p.Description = update.Description
	p.Status = update.Status
	p.StartDate = update.StartDate
	p.EndDate = update.EndDate
	p.UpdatedAt = time.Now()
	return nil
}

func (r *ProjectRepository) DeleteProject(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[id]; !ok {
		return errors.New("project not found")
	}
	delete(r.projects, id)
	delete(r.members, id)
	return nil
}

func (r *ProjectRepository) ListProjects() []*models.Project {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*models.Project
	for _, p := range r.projects {
		result = append(result, p)
	}
	return result
}

func (r *ProjectRepository) AddMember(projectID int, member *models.ProjectMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[projectID]; !ok {
		return errors.New("project not found")
	}
	member.JoinedAt = time.Now()
	r.members[projectID] = append(r.members[projectID], member)
	return nil
}

func (r *ProjectRepository) RemoveMember(projectID, userID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	members := r.members[projectID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[projectID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return errors.New("member not found")
}

func (r *ProjectRepository) ListMembers(projectID int) ([]*models.ProjectMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[projectID]; !ok {
		return nil, errors.New("project not found")
	}
	return r.members[projectID], nil
}
