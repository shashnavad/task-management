// services/file/models/file.go
package models

import "time"

type File struct {
	ID         int       `json:"id" db:"id"`
	FileName   string    `json:"file_name" db:"file_name"`
	FilePath   string    `json:"file_path" db:"file_path"`
	FileSize   int64     `json:"file_size" db:"file_size"`
	MimeType   string    `json:"mime_type" db:"mime_type"`
	ProjectID  *int      `json:"project_id" db:"project_id"`
	TaskID     *int      `json:"task_id" db:"task_id"`
	UploadedBy int       `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type FileUploadResponse struct {
	ID       int    `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	URL      string `json:"url"`
}
