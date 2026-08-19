package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
	"github.com/lib/pq"
)

var validStatuses = map[string]bool{
	"applied":      true,
	"interviewing": true,
	"offer":        true,
	"rejected":     true,
	"withdrawn":    true,
	"ghosted":      true,
}

type createApplicationEventRequest struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes"`
}

func CreateApplicationEventHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idStr := r.PathValue("id")
		jobApplicationID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req createApplicationEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if !validStatuses[req.Status] {
			writeError(w, http.StatusBadRequest, "invalid status")
			return
		}

		event, err := jobtracedb.CreateApplicationEvent(db, jobApplicationID, req.Status, req.Notes)

		if err != nil {
			// Check existence first so a bad jobApplicationID surfaces as sql.ErrNoRows,

			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "job application not found")
				return
			}

			// 23505 is Postgres's unique-violation code: application_events_unique_round
			// rejecting a duplicate round from a race, not an unexpected failure.

			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				writeError(w, http.StatusConflict, "round already recorded, please retry")
				return
			}

			log.Printf("create_application_event_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to create application event")
			return
		}

		writeJSON(w, http.StatusCreated, event)

	}
}

func DeleteApplicationEventHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobApplicationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		eventID, err := strconv.ParseInt(r.PathValue("event_id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid event_id")
			return
		}

		err = jobtracedb.DeleteApplicationEvent(db, jobApplicationID, eventID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "application event not found")
				return
			}

			log.Printf("delete_application_event_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to delete application event")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
