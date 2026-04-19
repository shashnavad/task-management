// services/reporting/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/reporting/handlers"
	"github.com/task-management/services/reporting/repository"
	"github.com/task-management/services/reporting/service"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	// Initialize database connections to all services
	authDB := repository.InitAuthDB()
	projectDB := repository.InitProjectDB()
	taskDB := repository.InitTaskDB()

	defer authDB.Close()
	defer projectDB.Close()
	defer taskDB.Close()

	// Initialize layers
	reportRepo := repository.NewReportRepository(authDB, projectDB, taskDB)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// Setup routes
	router := gin.Default()

	api := router.Group("/api/reports")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/dashboard", reportHandler.GetDashboard)
		api.GET("/project/:id/summary", reportHandler.GetProjectSummary)
		api.GET("/user/:id/productivity", reportHandler.GetUserProductivity)
		api.GET("/tasks/analytics", reportHandler.GetTaskAnalytics)
		api.GET("/export/csv", reportHandler.ExportToCSV)
		api.GET("/export/pdf", reportHandler.ExportToPDF)
	}

	logger.Info("Reporting service starting on port 8006")
	if err := http.ListenAndServe(":8006", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
