// services/project/models/project.go
package models

import "time"

type Project struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name" binding:"required"`
	Description string     `json:"description" db:"description"`
	OwnerID     int        `json:"owner_id" db:"owner_id"`
	Status      string     `json:"status" db:"status"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type ProjectMember struct {
	ID        int       `json:"id" db:"id"`
	ProjectID int       `json:"project_id" db:"project_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

type CreateProjectRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}
