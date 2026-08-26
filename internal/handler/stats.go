package handler

import (
	"database/sql"
	"log"
	"net/http"

	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
)

func GetFunnelSummaryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := jobtracedb.GetFunnelSummary(db)

		if err != nil {
			log.Printf("get_funnel_summary_handler failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to get funnel summary")
			return
		}

		writeJSON(w, http.StatusOK, summary)
	}
}
