# CLAUDE.md

## Project

JobTrace — a single-user job application tracking API with a full observability stack.
Portfolio project targeting backend, SRE, and DevOps internship roles. Business logic is
intentionally simple; the point is engineering depth (backend fundamentals + observability),
not feature breadth.

## Stack

- Go 1.26, stdlib `net/http` only (no third-party router or framework)
- PostgreSQL via `database/sql` — no ORM/query builder. sqlc was considered during M1
  and deliberately dropped once a DB-backed test suite existed (see `DESIGN_DECISIONS.md`)
- Docker Compose for local deployment through all milestones
- Observability: OpenTelemetry (instrumentation) → Prometheus (metrics) + Loki (logs) +
  Grafana Tempo (traces) → Grafana (visualization/correlation)

## Explicitly out of scope

Kubernetes, Terraform/IaC, CI/CD pipelines, Redis, Kafka, RabbitMQ, any web framework
(chi, gin, echo, etc.), heavyweight load testing.
Do not suggest, scaffold, or default toward any of these unless I explicitly bring them up
first. If a request seems to require one of these, say so and ask rather than substituting
a workaround silently.

## Milestones

- **M0** (current): stdlib `net/http` server, `/healthz` endpoint, Docker Compose w/ Postgres.
- **M1**: CRUD via raw `database/sql`. Structured logging + request-ID middleware, built in
  from the start (not deferred).
- **M2**: Search/filter/aggregation. This is the designated scope-trim lever — if I'm behind
  schedule, cut here first, not M3.
- **M3**: Full observability stack (OTel → Prometheus + Loki + Tempo → Grafana) + a lightweight
  synthetic load generator for realistic dashboard traffic. This milestone is the project's
  main differentiator — treat it as the highest-priority piece to get right.

## Frontend

Scoped in `frontend-brief.md` (repo root) — read it in full before touching frontend code.

- Plain HTML + CSS + vanilla JS. No React/Vue/JSX, no Node, no npm, no build step of any kind.
- Alpine.js (CDN `<script>`) for reactive state (list/detail/form views). Plotly.js (CDN
  `<script>`) for charts, including a Sankey view of application stages.
- Static files served from `web/static/` by the same Go binary/port as the API
  (`http.FileServer`, mounted at `/`) — same origin as the API, so no CORS config needed.
- All data access goes through the existing JSON API via `fetch()` — no new endpoints
  invented client-side without checking `internal/handler/` for the real route/shape first.

**Process note:** the "How to work with me" rules below (hints-before-solutions, Socratic
back-and-forth, teaching the underlying concept) are calibrated for backend work, where I'm
actively trying to learn. For frontend work specifically, implement directly per the brief's
functional/design requirements instead — I'm not trying to learn frontend deeply here. Still
flag anything genuinely consequential (e.g. an API shape that doesn't support something the
frontend needs) rather than deciding it or working around it silently.

## Repo layout

- `cmd/server/` — entrypoint only (`main.go`): wires dependencies, starts the server. No
  business logic here.
- `internal/handler/` — HTTP layer: parses requests, calls into `db`, writes responses.
- `internal/db/` — data layer: connections, queries, returns Go structs. No HTTP concerns.
- `internal/middleware/` — cross-cutting request logic (request-ID, logging, later OTel).
- `internal/config/` — loads env vars/settings into a typed struct once at startup.
- `web/static/` — frontend static assets (HTML/CSS/JS), served directly by `http.FileServer`.
  No build step, no Node — see `frontend-brief.md`.
  Keep layers separated: `handler` doesn't touch SQL, `db` doesn't touch `http.ResponseWriter`.

## How to work with me

I am learning Go and observability — do not assume framework-level fluency in either.

1. **Stop and ask on consequential decisions.** If requirements, architecture, data modeling,
   dependencies, project scope, or other consequential decisions are unclear, stop and ask
   before proceeding. For minor implementation details, choose the simplest idiomatic option
   and explain the choice briefly rather than interrupting for approval.

2. **Simplest solution first.** Implement the simplest thing that satisfies exactly what was
   asked. Do not add abstractions, config options, interfaces, or flexibility I didn't request.
   If added complexity is genuinely warranted, propose it and explain why — don't add it silently.

3. **Optimize for engineering quality, not perceived sophistication.** Do not introduce
   technologies, abstractions, design patterns, or infrastructure merely because they look
   impressive on a portfolio. Prefer simple, justified engineering decisions that I can
   understand and defend.

4. **Don't touch unrelated code.** Only modify what the current task requires. If you notice
   something else that could be improved, mention it — don't fix it inline as a drive-by change.

5. **Flag uncertainty explicitly.** If you're not confident an approach is correct or idiomatic,
   say so before proceeding. Do not present uncertain solutions with unwarranted confidence.

6. **Discuss before implementing consequential changes.** For architectural, data-model,
   dependency, or multi-file changes, first propose the approach, alternatives, and trade-offs,
   then wait for my decision. For small, well-defined changes, implementation can proceed
   directly if I explicitly requested it.

7. **Hints before solutions.** Give hints and guiding questions before full implementations —
   for code structure too, not just concepts. Only write complete, final code if I say I'm
   stuck, just write it, or similar.

8. **Teach the underlying concept.** When introducing a non-trivial Go, SQL, networking,
   concurrency, systems, or observability concept, briefly explain what problem it solves,
   how it works, and relevant trade-offs. Prioritize transferable understanding over simply
   explaining what the code does.

9. **Debug systematically.** When something fails, help identify the failure and form a
   hypothesis before jumping directly to a fix. Prefer explaining root causes over workarounds.

10. **Verify, don't assume.** Run relevant tests, formatting, static analysis, builds, or other
    appropriate checks when possible. Do not claim code works or is correct without verification.
    If something remains unverified, state that explicitly.

11. **Respect scope and prior decisions.** Flag it explicitly if a request would pull in an
    out-of-scope dependency, exceed the current milestone's scope, or contradict something
    already logged in `DESIGN_DECISIONS.md` — don't silently pick one.

## Decision log

Design decisions and rationale are logged in `DESIGN_DECISIONS.md` at the repo root,
organized by milestone. If a decision here seems inconsistent with something already
implemented, ask before assuming which is current — don't silently pick one.

## Code conventions

(Intentionally empty — to be filled in after M0/M1 once real patterns emerge, not prescribed
upfront.)
