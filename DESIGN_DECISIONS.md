## M0

### /healthz vs /readyz

Split into two endpoints instead of one combined check.

- `/healthz` (liveness) — confirms the process is running. No dependency checks.
- `/readyz` (readiness) — confirms the app can serve traffic, by pinging Postgres
  (2s timeout via `context.WithTimeout`).

**Why:** conflating the two is a known anti-pattern — if something uses a DB-checking
endpoint for liveness, a temporary DB blip causes unnecessary restarts. No orchestrator
here yet, but implementing the split correctly now avoids retrofitting later.

**Confirmed working:** `/readyz` returned 503 while Postgres was down, flipped to 200
right after `docker compose up`; `/healthz` stayed 200 throughout.

### sslmode=disable

DB connection string uses `sslmode=disable`. App and Postgres are both on `localhost` —
traffic never leaves the machine, so there's nothing in transit to encrypt. Revisit if
app and DB are ever split across machines (e.g. Pi self-hosting).

### CI pipeline from the first commit

Set up GitHub Actions (build, vet) before any real feature work, then grew it overtime as the project continued
to add a `gofmt -l` check and, once real tests existed, a Postgres service container.

**Why:** catches regressions and formatting drift on every push instead of relying on
remembering to run checks locally.

## M1

### `application_events_unique_round` partial unique index

`UNIQUE (job_application_id, round_number) WHERE status = 'interviewing'` on
`application_events` (migration `000003`).

**Why:** round_number is computed server-side (see below), but that's a read then a
write — two requests landing close together could read the same `MAX` and both try to
insert the same round. A `CHECK` only sees one row at a time so it can't catch this,
needed an index instead. Shows up as a Postgres unique-violation, which the handler
maps to a `409` via `errors.As` on the pq error code (`23505`).

There is a test function for this at application_events_test_go under handler folder.

### `round_number` computed server-side

Round number for a new `interviewing` event comes from the server
(`MAX(round_number) + 1` for that application), not the request body.

**Why:** feeds conversion stats and eventually a Sankey chart, both need rounds in the
right order. Letting the client send its own round number opens the door to gaps or
duplicates from a bad client. Server-side computation kills that; the unique index
above covers the race the computation alone can't.

### `updated_at` via trigger

`updated_at` is set by a `BEFORE UPDATE` trigger (`set_updated_at()`, migration
`000002`), not manually in each handler.

**Why:** app-managed `updated_at` is the same sync bug already avoided for `status`
elsewhere — one fact, two places, hoping every code path remembers to update both. A
trigger just makes it impossible to forget, no matter what touches the row.
