package db

import (
	"database/sql"
	"fmt"
	"time"
)

type ApplicationEvent struct {
	ID               int64
	JobApplicationID int64
	Status           string
	RoundNumber      *int
	OccurredAt       time.Time
	Notes            *string
}

func CreateApplicationEvent(db *sql.DB, jobApplicationID int64, status string, notes *string) (ApplicationEvent, error) {
	var exists int
	err := db.QueryRow(`SELECT 1 FROM job_applications WHERE id = $1`, jobApplicationID).Scan(&exists)
	if err != nil {
		return ApplicationEvent{}, fmt.Errorf("checking job application exists: %w", err)
	}

	// round_number is computed here, not accepted from the client, so it can't
	// be set out of sequence with the application's actual interview history.
	var roundNumber *int
	if status == "interviewing" {
		var next int
		err := db.QueryRow(
			`SELECT COALESCE(MAX(round_number), 0) + 1
			 FROM application_events
			 WHERE job_application_id = $1 AND status = 'interviewing'`,
			jobApplicationID,
		).Scan(&next)
		if err != nil {
			return ApplicationEvent{}, fmt.Errorf("computing next round number: %w", err)
		}
		roundNumber = &next
	}

	var event ApplicationEvent
	err = db.QueryRow(
		`INSERT INTO application_events (job_application_id, status, round_number, notes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, job_application_id, status, round_number, occurred_at, notes`,
		jobApplicationID, status, roundNumber, notes,
	).Scan(&event.ID, &event.JobApplicationID, &event.Status, &event.RoundNumber, &event.OccurredAt, &event.Notes)
	if err != nil {
		return ApplicationEvent{}, fmt.Errorf("inserting application event: %w", err)
	}

	return event, nil
}

func DeleteApplicationEvent(db *sql.DB, jobApplicationID, eventID int64) error {
	result, err := db.Exec(
		`DELETE FROM application_events WHERE id = $1 AND job_application_id = $2`,
		eventID, jobApplicationID,
	)
	if err != nil {
		return fmt.Errorf("deleting application event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
