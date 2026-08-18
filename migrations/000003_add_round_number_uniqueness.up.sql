CREATE UNIQUE INDEX application_events_unique_round
    ON application_events (job_application_id, round_number)
    WHERE status = 'interviewing';
