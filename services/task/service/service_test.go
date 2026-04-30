package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/task-management/services/task/models"
	"github.com/task-management/services/task/repository"
	_ "github.com/mattn/go-sqlite3"
)

type mockProducer struct {
	messages []producedMessage
}

type producedMessage struct {
	topic string
	value interface{}
}

func (m *mockProducer) Produce(topic string, value interface{}) error {
	m.messages = append(m.messages, producedMessage{topic: topic, value: value})
	return nil
}

func (m *mockProducer) Close() error {
	return nil
}

func newTestTaskService(t *testing.T) (*TaskService, *mockProducer, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	schema := `
CREATE TABLE tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	project_id INTEGER NOT NULL,
	assignee_id INTEGER,
	creator_id INTEGER NOT NULL,
	status TEXT,
	priority TEXT,
	due_date DATETIME,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
CREATE TABLE task_comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME NOT NULL
);`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := repository.NewTaskRepository(db)
	producer := &mockProducer{}
	return NewTaskService(repo, producer), producer, db
}

func TestCreateTask(t *testing.T) {
	service, producer, db := newTestTaskService(t)
	defer db.Close()

	task := &models.Task{
		ID:          101,
		Title:       "Write tests",
		Description: "create task service tests",
		ProjectID:   1,
		CreatorID:   1,
		CreatedBy:   1,
		Status:      "open",
		Priority:    "medium",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := service.CreateTask(task, 99); err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}

	tasks, err := service.GetTasks()
	if err != nil {
		t.Fatalf("GetTasks returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task in db, got %d", len(tasks))
	}
	if tasks[0].Title != "Write tests" {
		t.Fatalf("unexpected task title: got %q", tasks[0].Title)
	}
	if len(producer.messages) == 0 {
		t.Fatalf("expected produced events, got none")
	}
	if producer.messages[0].topic != "task.created" {
		t.Fatalf("expected first topic task.created, got %q", producer.messages[0].topic)
	}
}

func TestGetTask(t *testing.T) {
	service, _, db := newTestTaskService(t)
	defer db.Close()

	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO tasks (id, title, description, project_id, assignee_id, creator_id, status, priority, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "Existing task", "desc", 1, nil, 1, "open", "low", nil, now, now,
	)
	if err != nil {
		t.Fatalf("seed insert failed: %v", err)
	}

	task, err := service.GetTask(1)
	if err != nil {
		t.Fatalf("GetTask returned error: %v", err)
	}
	if task == nil {
		t.Fatalf("expected task, got nil")
	}
	if task.Title != "Existing task" {
		t.Fatalf("unexpected title: got %q", task.Title)
	}
}

func TestUpdateStatus(t *testing.T) {
	service, producer, db := newTestTaskService(t)
	defer db.Close()

	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO tasks (id, title, description, project_id, assignee_id, creator_id, status, priority, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "Status task", "desc", 1, nil, 1, "open", "low", nil, now, now,
	)
	if err != nil {
		t.Fatalf("seed insert failed: %v", err)
	}

	if err := service.UpdateStatus(1, "in_progress", 7); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	updated, err := service.GetTask(1)
	if err != nil {
		t.Fatalf("GetTask after update failed: %v", err)
	}
	if updated.Status != "in_progress" {
		t.Fatalf("status not updated: got %q", updated.Status)
	}
	if len(producer.messages) == 0 {
		t.Fatalf("expected status update event, got none")
	}
	if producer.messages[len(producer.messages)-1].topic != "task.status_updated" {
		t.Fatalf("expected task.status_updated event, got %q", producer.messages[len(producer.messages)-1].topic)
	}
}
