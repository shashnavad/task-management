package repository

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/task-management/services/reporting/models"
)

type ReportingRepository struct {
	mu sync.Mutex
	// In-memory data for demonstration
	projects   int
	tasks      int
	activities []models.ActivityItem
}

func NewReportingRepository() *ReportingRepository {
	return &ReportingRepository{
		projects:   10,
		tasks:      50,
		activities: []models.ActivityItem{{ID: 1, Type: "task", Description: "Task completed", UserName: "Alice", CreatedAt: time.Now()}},
	}
}

func (r *ReportingRepository) GetDashboardData() *models.DashboardData {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &models.DashboardData{
		TotalProjects:   r.projects,
		TotalTasks:      r.tasks,
		CompletedTasks:  30,
		PendingTasks:    15,
		OverdueTasks:    5,
		RecentActivity:  r.activities,
		TasksByStatus:   map[string]int{"completed": 30, "pending": 15, "overdue": 5},
		ProjectProgress: []models.ProjectProgress{{ProjectID: 1, ProjectName: "Demo Project", Progress: 0.8, TotalTasks: 10, CompletedTasks: 8}},
	}
}

func (r *ReportingRepository) GetProjectSummary(projectID int) *models.ProjectProgress {
	return &models.ProjectProgress{ProjectID: projectID, ProjectName: "Demo Project", Progress: 0.8, TotalTasks: 10, CompletedTasks: 8}
}

func (r *ReportingRepository) GetUserProductivity(userID int) map[string]interface{} {
	return map[string]interface{}{"user_id": userID, "tasks_completed": 12, "tasks_assigned": 15}
}

func (r *ReportingRepository) GetTaskAnalytics() map[string]interface{} {
	return map[string]interface{}{"total": 50, "completed": 30, "pending": 15, "overdue": 5}
}

func InitProjectDB() *ReportingRepository {
	return NewReportingRepository()
}

func InitTaskDB() *ReportingRepository {
	return NewReportingRepository()
}

// In-memory AuthRepository for demo purposes

type DBUser struct {
	ID          string
	Username    string
	PasswordHash string
	Email       string
	FullName    string
}

type AuthRepository struct {
	users map[string]*DBUser
	nextID int
}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{
		users: make(map[string]*DBUser),
		nextID: 1,
	}
}

func (r *AuthRepository) CreateUser(email, username, password, fullName string) (string, error) {
	for _, user := range r.users {
		if user.Username == username || user.Email == email {
			return "", errors.New("user already exists")
		}
	}
	id := strconv.Itoa(r.nextID)
	r.nextID++
	user := &DBUser{
		ID:          id,
		Username:    username,
		PasswordHash: password,
		Email:       email,
		FullName:    fullName,
	}
	r.users[id] = user
	return id, nil
}

func (r *AuthRepository) CheckUserExists(username, email string) (bool, error) {
	for _, user := range r.users {
		if user.Username == username || user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *AuthRepository) GetUserByUsername(username string) (*DBUser, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *AuthRepository) GetUserByID(userID string) (*DBUser, error) {
	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *AuthRepository) UpdateUser(userID, fullName, email string) error {
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.FullName = fullName
	user.Email = email
	return nil
}

func (r *AuthRepository) Close() error {
	return nil // No resources to clean up in-memory
}

func InitAuthDB() *AuthRepository {
	return NewAuthRepository()
}

func (r *ReportingRepository) Close() error {
	return nil // No resources to clean up in-memory
}

func NewReportRepository(authDB interface{}, projectDB interface{}, taskDB interface{}) *ReportingRepository {
	return NewReportingRepository()
}
