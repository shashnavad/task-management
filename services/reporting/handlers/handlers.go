package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/task-management/services/reporting/service"
)

type ReportingHandler struct {
	service *service.ReportingService
}

func NewReportingHandler(s *service.ReportingService) *ReportingHandler {
	return &ReportingHandler{service: s}
}

func (h *ReportingHandler) GetDashboard(c *gin.Context) {
	data := h.service.GetDashboardData()
	c.JSON(http.StatusOK, data)
}

func (h *ReportingHandler) GetProjectSummary(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	summary := h.service.GetProjectSummary(id)
	c.JSON(http.StatusOK, summary)
}

func (h *ReportingHandler) GetUserProductivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	productivity := h.service.GetUserProductivity(id)
	c.JSON(http.StatusOK, productivity)
}

func (h *ReportingHandler) GetTaskAnalytics(c *gin.Context) {
	analytics := h.service.GetTaskAnalytics()
	c.JSON(http.StatusOK, analytics)
}

// ExportToCSV handles exporting reports to CSV format (stub implementation)
func (h *ReportingHandler) ExportToCSV(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Export to CSV not implemented yet"})
}

// ExportToPDF handles exporting reports to PDF format (stub implementation)
func (h *ReportingHandler) ExportToPDF(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Export to PDF not implemented yet"})
}

// NewReportHandler creates a new ReportingHandler. Accepts interface{} for compatibility with main.go.
func NewReportHandler(s interface{}) *ReportingHandler {
	service, _ := s.(*service.ReportingService)
	return NewReportingHandler(service)
} 