package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
)

// Structs
type createJobApplicationRequest struct {
	Company   string  `json:"company"`
	RoleTitle string  `json:"role_title"`
	Notes     *string `json:"notes"`
}

type updateJobApplicationRequest struct {
	Company   *string `json:"company"`
	RoleTitle *string `json:"role_title"`
	Notes     *string `json:"notes"`
}

func CreateJobApplicationHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var req createJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.Company) == "" || strings.TrimSpace(req.RoleTitle) == "" {
			writeError(w, http.StatusBadRequest, "company and role_title are required")
			return
		}

		app, err := jobtracedb.CreateJobApplication(db, req.Company, req.RoleTitle, req.Notes)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "creation of job application failed")
			log.Printf("create_job_application_hander failed: %v", err)
			return

		}

		writeJSON(w, http.StatusCreated, app)

	}

}

func ListJobApplicationsHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var filter jobtracedb.JobApplicationFilter

		if status := r.URL.Query().Get("status"); status != "" {
			if !validStatuses[status] {
				writeError(w, http.StatusBadRequest, "invalid status")
				return
			}
			filter.Status = &status
		}

		if company := r.URL.Query().Get("company"); company != "" {
			filter.Company = &company
		}

		if roleTitle := r.URL.Query().Get("role_title"); roleTitle != "" {
			filter.RoleTitle = &roleTitle
		}

		if appliedAfterStr := r.URL.Query().Get("applied_after"); appliedAfterStr != "" {
			t, err := time.Parse("2006-01-02", appliedAfterStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid applied_after date, expected YYYY-MM-DD")
				return
			}
			filter.AppliedAfter = &t
		}

		if appliedBeforeStr := r.URL.Query().Get("applied_before"); appliedBeforeStr != "" {

			// In Golang, we use a reference date (must be 2006-01-02 to get YYYY-MM-DD ) to get the format that we want. Interesting....
			t, err := time.Parse("2006-01-02", appliedBeforeStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid applied_before date, expected YYYY-MM-DD")
				return
			}
			filter.AppliedBefore = &t
		}

		apps, err := jobtracedb.ListJobApplications(db, filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list job applications")
			log.Printf("list_job_applications_handler failed: %v", err)
			return
		}

		writeJSON(w, http.StatusOK, apps)

	}

}

func GetJobApplicationHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		detail, err := jobtracedb.GetJobApplication(db, id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "job application not found")
				return
			}

			log.Printf("get_job_application_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to get job application")
			return
		}

		writeJSON(w, http.StatusOK, detail)

	}

}

func UpdateJobApplicationHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req updateJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Check the fields. Company and Roletitle literally cannot be empty
		if req.Company != nil && strings.TrimSpace(*req.Company) == "" {
			writeError(w, http.StatusBadRequest, "company cannot be empty")
			return
		}
		if req.RoleTitle != nil && strings.TrimSpace(*req.RoleTitle) == "" {
			writeError(w, http.StatusBadRequest, "role_title cannot be empty")
			return
		}

		// Call the Database function to update the JobApplication!
		app, err := jobtracedb.UpdateJobApplication(db, id, req.Company, req.RoleTitle, req.Notes)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "job application not found")
				return
			}
			log.Printf("update_job_application_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to update job application")
			return
		}

		writeJSON(w, http.StatusOK, app)
	}

}

func DeleteJobApplicationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		err = jobtracedb.DeleteJobApplication(db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "job application not found")
				return
			}
			log.Printf("delete_job_application_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to delete job application")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
