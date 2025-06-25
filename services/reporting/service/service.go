package service

import (
	"github.com/task-management/services/reporting/models"
	"github.com/task-management/services/reporting/repository"
)

type ReportingService struct {
	repo *repository.ReportingRepository
}

func NewReportingService(repo *repository.ReportingRepository) *ReportingService {
	return &ReportingService{repo: repo}
}

func (s *ReportingService) GetDashboardData() *models.DashboardData {
	return s.repo.GetDashboardData()
}

func (s *ReportingService) GetProjectSummary(projectID int) *models.ProjectProgress {
	return s.repo.GetProjectSummary(projectID)
}

func (s *ReportingService) GetUserProductivity(userID int) map[string]interface{} {
	return s.repo.GetUserProductivity(userID)
}

func (s *ReportingService) GetTaskAnalytics() map[string]interface{} {
	return s.repo.GetTaskAnalytics()
}

// NewReportService creates a new ReportingService. Accepts interface{} for compatibility with main.go.
func NewReportService(repo interface{}) *ReportingService {
	r, _ := repo.(*repository.ReportingRepository)
	return NewReportingService(r)
}
