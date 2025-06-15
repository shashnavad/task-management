// services/reporting/models/report.go
package models

import "time"

type DashboardData struct {
	TotalProjects   int               `json:"total_projects"`
	TotalTasks      int               `json:"total_tasks"`
	CompletedTasks  int               `json:"completed_tasks"`
	PendingTasks    int               `json:"pending_tasks"`
	OverdueTasks    int               `json:"overdue_tasks"`
	RecentActivity  []ActivityItem    `json:"recent_activity"`
	TasksByStatus   map[string]int    `json:"tasks_by_status"`
	ProjectProgress []ProjectProgress `json:"project_progress"`
}

type ActivityItem struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectProgress struct {
	ProjectID      int     `json:"project_id"`
	ProjectName    string  `json:"project_name"`
	Progress       float64 `json:"progress"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
}
