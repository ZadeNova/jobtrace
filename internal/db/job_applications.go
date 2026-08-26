package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Use pointers when it is optional
type JobApplication struct {
	ID        int64     `json:"id"`
	Company   string    `json:"company"`
	RoleTitle string    `json:"role_title"`
	AppliedAt time.Time `json:"applied_at"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobApplicationSummary struct {
	ID              int64      `json:"id"`
	Company         string     `json:"company"`
	RoleTitle       string     `json:"role_title"`
	AppliedAt       time.Time  `json:"applied_at"`
	Notes           *string    `json:"notes"`
	CurrentStatus   *string    `json:"current_status"`
	CurrentRound    *int       `json:"current_round"`
	StatusUpdatedAt *time.Time `json:"status_updated_at"`
}

type JobApplicationDetail struct {
	JobApplicationSummary
	Events []ApplicationEvent `json:"events"`
}

type JobApplicationFilter struct {
	Status        *string
	Company       *string
	RoleTitle     *string
	AppliedAfter  *time.Time
	AppliedBefore *time.Time
}

func CreateJobApplication(db *sql.DB, company, roleTitle string, notes *string) (JobApplication, error) {

	tx, err := db.Begin()
	if err != nil {
		return JobApplication{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var app JobApplication

	err = tx.QueryRow(
		`INSERT INTO job_applications (company, role_title, notes)
		VALUES ($1, $2, $3)
		RETURNING id, company, role_title, applied_at, notes, created_at, updated_at`, company, roleTitle, notes,
	).Scan(&app.ID, &app.Company, &app.RoleTitle, &app.AppliedAt, &app.Notes, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		return JobApplication{}, fmt.Errorf("Inserting job application: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO application_events (job_application_id, status) VALUES ($1, 'applied')`, app.ID,
	)

	if err != nil {
		return JobApplication{}, fmt.Errorf("inserting initial application event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return JobApplication{}, fmt.Errorf("committing transaction: %w", err)
	}

	return app, nil

}

func ListJobApplications(db *sql.DB, filter JobApplicationFilter) ([]JobApplicationSummary, error) {

	query := `SELECT id, company, role_title, applied_at, notes, current_status, current_round, status_updated_at FROM job_application_summary WHERE 1=1`

	var args []any

	if filter.Status != nil {
		args = append(args, *filter.Status)
		query += fmt.Sprintf(" AND current_status = $%d", len(args))
	}

	if filter.Company != nil {
		args = append(args, "%"+*filter.Company+"%")
		query += fmt.Sprintf(" AND company ILIKE $%d", len(args))
	}

	if filter.RoleTitle != nil {
		args = append(args, "%"+*filter.RoleTitle+"%")
		query += fmt.Sprintf(" AND role_title ILIKE $%d", len(args))
	}

	if filter.AppliedAfter != nil {
		args = append(args, *filter.AppliedAfter)
		query += fmt.Sprintf(" AND applied_at >= $%d", len(args))
	}

	if filter.AppliedBefore != nil {
		args = append(args, *filter.AppliedBefore)
		query += fmt.Sprintf(" AND applied_at <= $%d", len(args))
	}

	query += " ORDER BY applied_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying job applications: %w", err)
	}
	defer rows.Close()

	apps := []JobApplicationSummary{}
	for rows.Next() {
		var app JobApplicationSummary
		if err := rows.Scan(&app.ID, &app.Company, &app.RoleTitle, &app.AppliedAt, &app.Notes, &app.CurrentStatus, &app.CurrentRound, &app.StatusUpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning job application: %w", err)
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating job applications: %w", err)
	}

	return apps, nil
}

func GetJobApplication(db *sql.DB, id int64) (JobApplicationDetail, error) {
	var detail JobApplicationDetail

	err := db.QueryRow(
		`SELECT id, company, role_title, applied_at, notes, current_status, current_round, status_updated_at
		 FROM job_application_summary WHERE id = $1`, id,
	).Scan(&detail.ID, &detail.Company, &detail.RoleTitle, &detail.AppliedAt, &detail.Notes, &detail.CurrentStatus, &detail.CurrentRound, &detail.StatusUpdatedAt)
	if err != nil {
		return JobApplicationDetail{}, fmt.Errorf("querying job application: %w", err)
	}

	rows, err := db.Query(
		`SELECT id, job_application_id, status, round_number, occurred_at, notes
		 FROM application_events WHERE job_application_id = $1 ORDER BY occurred_at`, id,
	)
	if err != nil {
		return JobApplicationDetail{}, fmt.Errorf("querying application events: %w", err)
	}
	defer rows.Close()

	events := []ApplicationEvent{}
	for rows.Next() {
		var e ApplicationEvent
		if err := rows.Scan(&e.ID, &e.JobApplicationID, &e.Status, &e.RoundNumber, &e.OccurredAt, &e.Notes); err != nil {
			return JobApplicationDetail{}, fmt.Errorf("scanning application event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return JobApplicationDetail{}, fmt.Errorf("iterating application events: %w", err)
	}

	detail.Events = events
	return detail, nil
}

func UpdateJobApplication(db *sql.DB, id int64, company, roleTitle, notes *string) (JobApplication, error) {

	var app JobApplication

	err := db.QueryRow(
		`UPDATE job_applications
		SET company = COALESCE($1, company),
		role_title = COALESCE($2, role_title),
		notes = COALESCE($3, notes)
		WHERE id = $4
		RETURNING id, company, role_title, applied_at, notes, created_at, updated_at`, company, roleTitle, notes, id,
	).Scan(&app.ID, &app.Company, &app.RoleTitle, &app.AppliedAt, &app.Notes, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		return JobApplication{}, fmt.Errorf("updating job application: %w", err)
	}

	return app, nil

}

func DeleteJobApplication(db *sql.DB, id int64) error {
	result, err := db.Exec(`DELETE FROM job_applications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting job application: %w", err)
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
