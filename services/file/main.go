// services/file/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/file/handlers"
	"github.com/task-management/services/file/repository"
	"github.com/task-management/services/file/service"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize database
	db := repository.InitDB()
	defer db.Close()

	// Initialize layers
	fileRepo := repository.NewFileRepository(db)
	fileService := service.NewFileService(fileRepo)
	fileHandler := handlers.NewFileHandler(fileService)

	// Setup routes
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	api := router.Group("/api/files")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/upload", fileHandler.UploadFile)
		api.GET("/:id", fileHandler.GetFile)
		api.GET("/:id/download", fileHandler.DownloadFile)
		api.DELETE("/:id", fileHandler.DeleteFile)
		api.GET("/project/:projectId", fileHandler.GetProjectFiles)
		api.GET("/task/:taskId", fileHandler.GetTaskFiles)
	}

	logger.Info("File service starting on port 8004")
	if err := http.ListenAndServe(":8004", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
