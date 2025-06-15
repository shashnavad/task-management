// services/reporting/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-management/reporting/handlers"
	"github.com/task-management/reporting/repository"
	"github.com/task-management/reporting/service"
)

func main() {
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
	api.Use(authMiddleware())
	{
		api.GET("/dashboard", reportHandler.GetDashboard)
		api.GET("/project/:id/summary", reportHandler.GetProjectSummary)
		api.GET("/user/:id/productivity", reportHandler.GetUserProductivity)
		api.GET("/tasks/analytics", reportHandler.GetTaskAnalytics)
		api.GET("/export/csv", reportHandler.ExportToCSV)
		api.GET("/export/pdf", reportHandler.ExportToPDF)
	}

	log.Println("Reporting service starting on port 8006")
	log.Fatal(http.ListenAndServe(":8006", router))
}
