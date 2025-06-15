// shared/events/events.go
package events

type TaskCreatedEvent struct {
	TaskID     int    `json:"task_id"`
	ProjectID  int    `json:"project_id"`
	AssigneeID *int   `json:"assignee_id"`
	Title      string `json:"title"`
	CreatedBy  int    `json:"created_by"`
	Timestamp  int64  `json:"timestamp"`
}

type TaskUpdatedEvent struct {
	TaskID    int    `json:"task_id"`
	Status    string `json:"status"`
	UpdatedBy int    `json:"updated_by"`
	Timestamp int64  `json:"timestamp"`
}

type ProjectCreatedEvent struct {
	ProjectID int    `json:"project_id"`
	Name      string `json:"name"`
	OwnerID   int    `json:"owner_id"`
	Timestamp int64  `json:"timestamp"`
}
