// services/task/models/task.go
package models

import "time"

type Task struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title" binding:"required"`
	Description string     `json:"description" db:"description"`
	ProjectID   int        `json:"project_id" db:"project_id" binding:"required"`
	AssigneeID  *int       `json:"assignee_id" db:"assignee_id"`
	CreatorID   int        `json:"creator_id" db:"creator_id"`
	Status      string     `json:"status" db:"status"`
	Priority    string     `json:"priority" db:"priority"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type TaskComment struct {
	ID        int       `json:"id" db:"id"`
	TaskID    int       `json:"task_id" db:"task_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	ProjectID   int        `json:"project_id" binding:"required"`
	AssigneeID  *int       `json:"assignee_id"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}
