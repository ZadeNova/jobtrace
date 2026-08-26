package db

import (
	"database/sql"
	"fmt"
)

type FunnelSummary struct {
	TotalApplications     int     `json:"total_applications"`
	Interviewed           int     `json:"interviewed"`
	InterviewRate         float64 `json:"interview_rate"`
	Offered               int     `json:"offered"`
	OfferRateOfTotal      float64 `json:"offer_rate_of_total"`
	OfferRateOfInterviews float64 `json:"offer_rate_of_interviews"`
}

type FlowEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Count int    `json:"count"`
}

func GetFunnelSummary(db *sql.DB) (FunnelSummary, error) {
	var s FunnelSummary

	err := db.QueryRow(
		`SELECT
			(SELECT COUNT(*) FROM job_applications),
			(SELECT COUNT(DISTINCT job_application_id) FROM application_events WHERE status = 'interviewing'),
			(SELECT COUNT(DISTINCT job_application_id) FROM application_events WHERE status = 'offer')`,
	).Scan(&s.TotalApplications, &s.Interviewed, &s.Offered)
	if err != nil {
		return FunnelSummary{}, fmt.Errorf("querying funnel summary: %w", err)
	}

	// Prevent the zero division eorror
	if s.TotalApplications > 0 {
		s.InterviewRate = float64(s.Interviewed) / float64(s.TotalApplications) * 100
		s.OfferRateOfTotal = float64(s.Offered) / float64(s.TotalApplications) * 100
	}
	if s.Interviewed > 0 {
		s.OfferRateOfInterviews = float64(s.Offered) / float64(s.Interviewed) * 100
	}

	return s, nil
}

func GetApplicationFlow(db *sql.DB) ([]FlowEdge, error) {
	rows, err := db.Query(`
		WITH labeled_events AS (
			SELECT
				job_application_id,
				occurred_at,
				CASE
					WHEN status = 'interviewing' THEN 'interviewing_round_' || round_number
					ELSE status
				END AS stage
			FROM application_events
		)
		SELECT from_stage, to_stage, COUNT(*) AS count
		FROM (
			SELECT
				LAG(stage) OVER (PARTITION BY job_application_id ORDER BY occurred_at) AS from_stage,
				stage AS to_stage
			FROM labeled_events
		) transitions
		WHERE from_stage IS NOT NULL
		GROUP BY from_stage, to_stage
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying application flow: %w", err)
	}
	defer rows.Close()

	edges := []FlowEdge{}
	for rows.Next() {
		var e FlowEdge
		if err := rows.Scan(&e.From, &e.To, &e.Count); err != nil {
			return nil, fmt.Errorf("scanning flow edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating flow edges: %w", err)
	}

	return edges, nil
}
