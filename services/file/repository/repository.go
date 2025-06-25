package repository

import (
	"database/sql"
	"github.com/task-management/services/file/models"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) SaveFile(file *models.File) error {
	stmt, err := r.db.Prepare(`INSERT INTO files (file_name, file_path, file_size, mime_type, project_id, task_id, uploaded_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(file.FileName, file.FilePath, file.FileSize, file.MimeType, file.ProjectID, file.TaskID, file.UploadedBy, file.CreatedAt)
	return err
}

func (r *FileRepository) GetFileByID(id int) (*models.File, error) {
	file := &models.File{}
	row := r.db.QueryRow(`SELECT id, file_name, file_path, file_size, mime_type, project_id, task_id, uploaded_by, created_at FROM files WHERE id = ?`, id)
	err := row.Scan(&file.ID, &file.FileName, &file.FilePath, &file.FileSize, &file.MimeType, &file.ProjectID, &file.TaskID, &file.UploadedBy, &file.CreatedAt)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (r *FileRepository) DeleteFile(id int) error {
	stmt, err := r.db.Prepare(`DELETE FROM files WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (r *FileRepository) GetFilesByProjectID(projectID int) ([]*models.File, error) {
	rows, err := r.db.Query(`SELECT id, file_name, file_path, file_size, mime_type, project_id, task_id, uploaded_by, created_at FROM files WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []*models.File
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(&file.ID, &file.FileName, &file.FilePath, &file.FileSize, &file.MimeType, &file.ProjectID, &file.TaskID, &file.UploadedBy, &file.CreatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *FileRepository) GetFilesByTaskID(taskID int) ([]*models.File, error) {
	rows, err := r.db.Query(`SELECT id, file_name, file_path, file_size, mime_type, project_id, task_id, uploaded_by, created_at FROM files WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []*models.File
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(&file.ID, &file.FileName, &file.FilePath, &file.FileSize, &file.MimeType, &file.ProjectID, &file.TaskID, &file.UploadedBy, &file.CreatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func InitDB() *sql.DB {
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/file_service")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to file service database")
	return db
} 