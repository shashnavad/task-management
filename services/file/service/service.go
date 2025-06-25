package service

// import "github.com/task-management/services/file/models" // Remove if not used directly

import "github.com/task-management/services/file/models"

type FileRepository interface {
	SaveFile(file *models.File) error
	GetFileByID(id int) (*models.File, error)
	DeleteFile(id int) error
	GetFilesByProjectID(projectID int) ([]*models.File, error)
	GetFilesByTaskID(taskID int) ([]*models.File, error)
}

type FileService struct {
	repo FileRepository
}

func NewFileService(repo FileRepository) *FileService {
	return &FileService{repo: repo}
}

func (s *FileService) UploadFile(file *models.File) error {
	return s.repo.SaveFile(file)
}

func (s *FileService) GetFile(id int) (*models.File, error) {
	return s.repo.GetFileByID(id)
}

func (s *FileService) DeleteFile(id int) error {
	return s.repo.DeleteFile(id)
}

func (s *FileService) GetProjectFiles(projectID int) ([]*models.File, error) {
	return s.repo.GetFilesByProjectID(projectID)
}

func (s *FileService) GetTaskFiles(taskID int) ([]*models.File, error) {
	return s.repo.GetFilesByTaskID(taskID)
}

// Add your service methods here, e.g.:
// func (s *FileService) UploadFile(...) error { ... }
