package handlers

import (
	"encoding/json"
	"net/http"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"
)

// -- Reports Handlers --

func CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req models.Report
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if req.ReporterID == 0 || req.ReportedUserID == 0 || req.Reason == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing required fields")
		return
	}

	id, err := database.CreateReport(database.DBConn, &req)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to create report: "+err.Error())
		return
	}
	req.ID = id
	utils.WriteJSON(w, http.StatusCreated, req)
}

func ListReportsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	reports, err := database.GetAllReports(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, reports)
}
