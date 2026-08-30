# Phase 3 Plan

Sequencing for M3 (observability) and the work that follows it. Not fixed — expect this
to change if scope expands or contracts, or a new design decision changes the shape of a
later step.

## Order of work

1. **Prometheus + Grafana** — expose `/metrics`, stand up both containers, first real
   dashboard. Paired together as one checkpoint since Prometheus alone has no real UI
   worth using; Grafana is what makes the metrics useful.

2. **Loki** — ship the app's `slog` logs to it via **Grafana Alloy** (the log-shipping
   agent; originally planned as Promtail, swapped out after Promtail's EOL). Switch the
   app's log output to JSON so Alloy can parse fields into Loki labels, then add Loki as
   a second Grafana datasource. `request_id` is the queryable correlation label for now;
   `trace_id`/`span_id` correlation is deferred to checkpoint 3, once real spans exist.
   No changes needed to checkpoint 1's Prometheus/Grafana pieces.

3. **OTel + Tempo** — instrument the app with spans, stand up Tempo, add as a third
   Grafana datasource. Paired together since neither produces anything visible alone —
   instrumentation with nowhere to send spans, or a trace store with nothing sent to it,
   are both dead ends on their own. This step will need `context.Context` added to
   `internal/db` function signatures and a switch to `QueryContext`/`ExecContext` — a
   real, multi-file change to flag properly when we get there, not decide silently.

**M3 is done once all three checkpoints above work end to end.**

4. **Frontend** — build per `frontend-brief.md`, so the project is actually usable
   day to day. Real personal usage from here on generates real traffic through the
   observability stack built in steps 1-3, instead of relying on synthetic filler.

5. **Synthetic load generator** — deliberately last. Targets a separate database
   (e.g. `jobtrace_demo`), never the real `jobtrace` database — its only purpose is
   producing realistic-looking dashboard/demo traffic, not real usage. Building it last
   means it's a polish/demo tool on top of a project that already works, not something
   needed just to see the dashboards populated.

## Note on `CLAUDE.md`

`CLAUDE.md`'s M3 description currently lists the load generator as part of M3 itself.
This plan deliberately pulls it out and pushes it past the frontend — worth reconciling
in `CLAUDE.md` at some point, flagging rather than silently leaving the two documents
disagreeing.
