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
