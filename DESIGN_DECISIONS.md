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
