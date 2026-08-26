# JobTrace Frontend — Brief for Claude Code

This is a scoping document, not a build command yet. Read this fully before writing any
frontend code. It supersedes nothing in `CLAUDE.md` for the backend — this brief only
covers the frontend, and explicitly relaxes one of `CLAUDE.md`'s rules for this slice of
work (see "Process notes" below).

## Context

JobTrace's backend (Go, `net/http`, Postgres) already exists. This brief scopes a
frontend so I can actually use the tool day-to-day as I start applying to jobs, and
so the tool has something visual/usable beyond curl.

## Stack decision (final — do not substitute)

- **Plain HTML + CSS + vanilla JS.** No React, no Vue, no JSX.
- **Alpine.js**, loaded via CDN `<script>` tag. Chosen over raw JS specifically to
  handle reactive state (`x-data`, `x-for`, `x-show`, `x-model`) for the list/detail/form
  views without hand-rolled DOM-diffing and view-switching logic.
- **Plotly.js**, loaded via CDN `<script>` tag, for charts — including a Sankey funnel
  view of application stages (applied → interview → offer, etc.).
- **No Node, no npm, no build step of any kind.** This is a hard constraint, not a
  preference — do not introduce a bundler, TypeScript compiler, or any tool that requires
  `node_modules`.
- No CSS framework (no Tailwind). Hand-written CSS with a deliberate palette/type system.

## Architecture

- Static files (`index.html`, `style.css`, `app.js` or similar) are served by the
  **same Go binary and port** as the API — e.g. `http.FileServer` for `/`, existing
  handlers under `/api/...`. Same origin, so `fetch()` calls need no CORS config.
- All data operations go through **existing JSON API endpoints** via `fetch()` —
  GET for list/detail, POST for create, PUT/PATCH for update, DELETE for delete.
  Check `internal/handler/` for the actual current route names/shapes rather than
  assuming — don't invent endpoint paths.
- Charts consume the M2 aggregation endpoint's output directly — that endpoint's
  job is to hand back pre-aggregated numbers (status breakdown, interview %, etc.),
  so chart-side logic should stay thin.

## Scope (v1)

- List view: table of applications (company, role, status, date, etc.)
- Detail view: full record for one application
- Create/edit form
- Delete (with confirmation)
- Status breakdown chart + Sankey funnel of application stages
- Responsive down to mobile; visible keyboard focus states

## Explicit non-goals

- **Grafana is not part of this.** Grafana stays scoped entirely to M3's observability
  stack (metrics/logs/traces about the API's runtime behavior). No job-application data
  flows through Grafana. Don't conflate the two.
- No user accounts/auth — single-user tool.
- No animation library — use plain CSS `transition`/`@keyframes` unless a specific
  animation genuinely can't be done in CSS (flag it if so, don't just add a library).

## Suggested file split

Once this crosses a couple hundred lines, split rather than leaving it in one file:
`index.html`, `style.css`, `api.js` (fetch wrappers), `app.js` (Alpine state + rendering),
`charts.js` (Plotly setup).

## Design direction

I'll provide the actual visual direction (palette, type pairing, layout concept) separately
before you style anything — don't default to generic AI-look patterns (cream + terracotta,
near-black + neon accent, hairline-rule broadsheet layout) in the meantime.

## Process notes for Claude Code

Unlike the backend work, I'm not trying to learn frontend deeply here — I'll give you the
design direction and functional requirements, and want you to just implement it well,
without the usual hints-before-solutions/Socratic back-and-forth `CLAUDE.md` asks for on
the backend. Still flag anything genuinely consequential (e.g. if an API shape doesn't
support something the frontend needs) rather than working around it silently.
