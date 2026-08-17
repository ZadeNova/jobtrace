CREATE TABLE job_applications (
    id         BIGSERIAL PRIMARY KEY,
    company    TEXT NOT NULL,
    role_title TEXT NOT NULL,
    applied_at DATE NOT NULL DEFAULT CURRENT_DATE,
    notes      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE application_events (
    id                  BIGSERIAL PRIMARY KEY,
    job_application_id  BIGINT NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    status              TEXT NOT NULL CHECK (status IN
                             ('applied', 'interviewing', 'offer', 'rejected', 'withdrawn', 'ghosted')),
    round_number        INT,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    notes               TEXT,
    CHECK (
        (status = 'interviewing' AND (round_number IS NULL OR round_number > 0))
        OR (status <> 'interviewing' AND round_number IS NULL)
    )
);

CREATE INDEX idx_application_events_job_application_id_occurred_at
    ON application_events (job_application_id, occurred_at DESC);

CREATE VIEW job_application_summary AS
SELECT
    ja.id,
    ja.company,
    ja.role_title,
    ja.applied_at,
    ja.notes,
    ce.status        AS current_status,
    ce.round_number  AS current_round,
    ce.occurred_at   AS status_updated_at
FROM job_applications ja
LEFT JOIN LATERAL (
    SELECT status, round_number, occurred_at
    FROM application_events ae
    WHERE ae.job_application_id = ja.id
    ORDER BY occurred_at DESC
    LIMIT 1
) ce ON true;
