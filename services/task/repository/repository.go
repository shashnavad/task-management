package repository

import (
	"database/sql"
	"github.com/task-management/services/task/models"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetTasks() ([]*models.Task, error) {
	rows, err := r.db.Query(`SELECT id, title, description, project_id, assignee_id, creator_id, status, priority, due_date, created_at, updated_at FROM tasks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []*models.Task
	for rows.Next() {
		t := &models.Task{}
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTask(id int) (*models.Task, error) {
	t := &models.Task{}
	row := r.db.QueryRow(`SELECT id, title, description, project_id, assignee_id, creator_id, status, priority, due_date, created_at, updated_at FROM tasks WHERE id = ?`, id)
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) CreateTask(task *models.Task) error {
	stmt, err := r.db.Prepare(`INSERT INTO tasks (title, description, project_id, assignee_id, creator_id, status, priority, due_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.Title, task.Description, task.ProjectID, task.AssigneeID, task.CreatorID, task.Status, task.Priority, task.DueDate, time.Now(), time.Now())
	return err
}

func (r *TaskRepository) UpdateTask(task *models.Task) error {
	stmt, err := r.db.Prepare(`UPDATE tasks SET title=?, description=?, project_id=?, assignee_id=?, status=?, priority=?, due_date=?, updated_at=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.Title, task.Description, task.ProjectID, task.AssigneeID, task.Status, task.Priority, task.DueDate, time.Now(), task.ID)
	return err
}

func (r *TaskRepository) DeleteTask(id int) error {
	stmt, err := r.db.Prepare(`DELETE FROM tasks WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (r *TaskRepository) AssignTask(id int, assigneeID int) error {
	stmt, err := r.db.Prepare(`UPDATE tasks SET assignee_id=?, updated_at=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(assigneeID, time.Now(), id)
	return err
}

func (r *TaskRepository) UpdateStatus(id int, status string) error {
	stmt, err := r.db.Prepare(`UPDATE tasks SET status=?, updated_at=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(status, time.Now(), id)
	return err
}

func (r *TaskRepository) AddComment(comment *models.TaskComment) error {
	stmt, err := r.db.Prepare(`INSERT INTO task_comments (task_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(comment.TaskID, comment.UserID, comment.Content, time.Now())
	return err
}

// InitDB initializes the database connection for the task service
func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./task.db")
	if err != nil {
		panic(err)
	}
	// Optionally, you can run migrations or ensure tables exist here
	return db
} 