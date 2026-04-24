package handler

import (
	"log"
	"net/http"

	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// ReportHandler holds HTTP handlers for report endpoints.
type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// HandleUserReport handles GET /api/admin/reports/users
func (h *ReportHandler) HandleUserReport(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	data, err := h.reportService.GetUserReport(r.Context(), from, to)
	if err != nil {
		log.Printf("ERROR: GetUserReport failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate user report")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleRevenueReport handles GET /api/admin/reports/revenue
func (h *ReportHandler) HandleRevenueReport(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	data, err := h.reportService.GetRevenueReport(r.Context(), from, to)
	if err != nil {
		log.Printf("ERROR: GetRevenueReport failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate revenue report")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleDrawReport handles GET /api/admin/reports/draws
func (h *ReportHandler) HandleDrawReport(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	data, err := h.reportService.GetDrawReport(r.Context(), from, to)
	if err != nil {
		log.Printf("ERROR: GetDrawReport failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate draw report")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleCharityReport handles GET /api/admin/reports/charities
func (h *ReportHandler) HandleCharityReport(w http.ResponseWriter, r *http.Request) {
	data, err := h.reportService.GetCharityReport(r.Context())
	if err != nil {
		log.Printf("ERROR: GetCharityReport failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate charity report")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}
