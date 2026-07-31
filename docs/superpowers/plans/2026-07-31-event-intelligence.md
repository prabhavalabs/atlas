# Event correlation, timelines, and reports implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Turn versioned observations into explainable canonical events, meaningful cited timelines, and publishable deterministic reports with optional validated model rendering.

**Architecture:** Deterministic, versioned rules retrieve and score correlation candidates. Event revisions and evidence links are immutable. Material-change extraction creates reviewable timeline candidates and a structured evidence bundle. Publication is a human command.

**Tech stack:** Go, PostGIS, PostgreSQL full-text/trigram, pgx/sqlc, optional OpenAI-compatible HTTP renderer.

### Task 1: Implement versioned event taxonomy and rule configuration

**Files:**

- Create: `configs/taxonomy.yaml`
- Create: `configs/correlation.yaml`
- Create: `configs/priority.yaml`
- Create: `internal/event/rules.go`
- Create: `internal/event/rules_test.go`

**Steps:**

1. Define stable event types, compatible source categories, lifecycle transitions, material fields, spatial/time windows, correlation weights/bands, and public priority rules.
2. Validate configs at startup for missing types, invalid ranges, overlapping bands, negative weights, and unknown source mappings.
3. Hash/version the effective configuration and store that version on decisions/revisions.
4. Write golden tests for representative earthquake, cyclone, flood, wildfire, and unknown event inputs.
5. Ensure any missing optional value remains unknown and cannot contribute a numeric zero.

### Task 2: Build correlation candidate retrieval and scoring

**Files:**

- Create: `migrations/00007_event_correlation.sql`
- Create: `queries/event/correlation.sql`
- Create: `internal/event/correlation/features.go`
- Create: `internal/event/correlation/service.go`
- Create: `internal/event/correlation/service_test.go`
- Create: `internal/event/correlation/integration_test.go`

**Steps:**

1. Add event aliases, event links, candidate decisions, geometry versions, and score feature JSON with constraints/indexes.
2. Test exact alias/cross-reference, compatible-type filter, space/time retrieval, distance at antimeridian, title/place similarity, magnitude/track consistency, and missing values.
3. Score candidates with versioned deterministic weights and return `auto_link`, `new_event`, or `review` bands.
4. In one transaction create/link event, revision, decision record, and follow-up material-change job.
5. Make replay idempotent and store every feature/threshold so a decision can be explained later.

### Task 3: Calibrate on the golden event corpus

**Files:**

- Create: `testdata/events/golden/manifest.yaml`
- Create: `testdata/events/golden/*.json`
- Create: `internal/event/correlation/golden_test.go`
- Create: `cmd/atlas/evaluate.go`

**Steps:**

1. Add cases listed in `docs/14-testing-and-quality.md`, with expected link/new/review decisions and licensed/redacted fixtures.
2. Implement evaluation output for true/false link, duplicate, review rate, and per-feature decision explanation.
3. Set initial thresholds to favor ambiguous review over false merge.
4. Require reviewed golden diff for every rules change; CI fails on unapproved expectation drift.
5. Record calibration dataset/version and known coverage gaps.

### Task 4: Implement lifecycle, priority, and revision invariants

**Files:**

- Create: `internal/event/lifecycle.go`
- Create: `internal/event/priority.go`
- Create: `internal/event/revisions.go`
- Create: `internal/event/lifecycle_test.go`
- Create: `internal/event/priority_test.go`
- Create: `internal/event/revisions_integration_test.go`

**Steps:**

1. Write transition-table tests for observed/active/monitoring/closed/rejected/reactivated plus invalid/test source behavior.
2. Add type-specific inactivity policies and downgrade hysteresis; missing fetches never close/downgrade.
3. Calculate priority only from local stored facts and source class; separate it from source severity and confidence.
4. Create revisions only on material canonical changes, with actor/rule version/reason and previous revision.
5. Property-test that link/replay order cannot produce a reassuring state solely from an error/unknown.

### Task 5: Implement reviewable merge and split commands

**Files:**

- Create: `internal/event/commands/merge.go`
- Create: `internal/event/commands/split.go`
- Create: `internal/event/commands/merge_split_integration_test.go`
- Create: `internal/adminapi/correlation.go`

**Steps:**

1. Test merge of links/aliases/timeline/article/report references, slug redirect, revision/audit, and repeat idempotency.
2. Test split moving selected observations to a new event, recomputing canonical fields, and preserving public history/correction note.
3. Require current revision and written reason; conflict returns 409 without partial changes.
4. Keep operations reversible by storing a decision bundle; implement undo for an unpublished merge in MVP.
5. Queue report/timeline/read-model recomputation transactionally.

### Task 6: Extract material fact changes

**Files:**

- Create: `internal/editorial/facts/model.go`
- Create: `internal/editorial/facts/extract.go`
- Create: `internal/editorial/facts/extract_test.go`
- Create: `internal/editorial/facts/golden_test.go`

**Steps:**

1. Define typed fact keys, value/range/unit, valid time, status, confidence, and evidence version.
2. Test first observation, numeric revision, unit conversion, new geometry/track, alert escalation/downgrade, correction/retraction, conflict, and formatting-only no-op.
3. Use event-type materiality rules and preserve before/after; do not overwrite conflicting facts.
4. Produce deterministic stable hashes so duplicate deliveries do not create candidates.
5. Golden-test human-readable change titles while keeping fact data structured.

### Task 7: Create cited timeline candidates and publication

**Files:**

- Create: `migrations/00008_editorial.sql`
- Create: `queries/editorial/timeline.sql`
- Create: `internal/editorial/timeline/service.go`
- Create: `internal/editorial/timeline/service_test.go`
- Create: `internal/editorial/timeline/integration_test.go`
- Create: `internal/adminapi/timeline.go`

**Steps:**

1. Add timeline entries, revisions, citations, publication/verification/materiality fields with a constraint requiring citations for factual publication.
2. Cluster corroborating fact changes and retain conflicts. Generate candidate title/body from deterministic templates.
3. Auto-publish only explicitly approved low-risk official update classes; default to review.
4. Admin edits preserve evidence links; suppressing evidence blocks/retracts dependent draft publication.
5. Test chronological/as-of ordering, dense grouping, citation navigation, correction, withdrawal, and replay.

### Task 8: Add articles, canonicalization, and deterministic ranking

**Files:**

- Create: `queries/editorial/articles.sql`
- Create: `internal/editorial/articles/url.go`
- Create: `internal/editorial/articles/cluster.go`
- Create: `internal/editorial/articles/rank.go`
- Create: `internal/editorial/articles/articles_test.go`

**Steps:**

1. Test URL canonicalization without stripping identity-critical parameters; normalized-title hash; publication time validation; future/undated quarantine.
2. Cluster exact/near headlines and explicit source links; retain independent publisher count.
3. Rank by source tier, event relevance, corroboration, freshness, and material change using versioned weights.
4. Store metadata/excerpt only under source policy; preserve original and canonical URLs.
5. Ensure popularity/news volume cannot change event severity/priority.

### Task 9: Build structured evidence bundles and template reports

**Files:**

- Create: `queries/editorial/reports.sql`
- Create: `internal/editorial/report/schema.go`
- Create: `internal/editorial/report/evidence.go`
- Create: `internal/editorial/report/template.go`
- Create: `internal/editorial/report/report_test.go`
- Create: `internal/editorial/report/integration_test.go`

**Steps:**

1. Define report sections/claims/citations from `docs/08-reports-and-ai.md` and validate unknown/conflict semantics.
2. Select bounded current facts plus material timeline and recent authoritative response items; record exact evidence hash.
3. Render concise deterministic prose with citation locators and explicit “no verified information” unknowns.
4. Store draft/version separately; publishing atomically sets current version and creates audit/publication revision.
5. Test sparse evidence, conflicts, report update, unchanged regeneration, suppressed evidence, and corrected publication.

### Task 10: Implement optional validated renderers

**Files:**

- Create: `internal/editorial/report/renderer.go`
- Create: `internal/editorial/report/ollama.go`
- Create: `internal/editorial/report/openai_compatible.go`
- Create: `internal/editorial/report/validate.go`
- Create: `internal/editorial/report/renderer_test.go`
- Create: `configs/report-style.yaml`

**Steps:**

1. Implement `template`, `ollama`, and `openai-compatible` behind one interface; default to template.
2. Use strict JSON schema, model allowlist, time/token/request budgets, no tools/network/browsing, and bounded evidence.
3. Test invalid JSON, unknown citation, uncited claim, changed number/unit/date, unsafe imperative, timeout/429, and provider outage.
4. Fall back to template without losing draft workflow; record safe diagnostics/model/prompt/token/evidence metadata.
5. Require admin trigger per report/section and local hard monthly budget; no automatic scheduled model calls.

### Task 11: Add public read models, search, and cache validators

**Files:**

- Create: `migrations/00009_public_search.sql`
- Create: `queries/publicapi/events.sql`
- Create: `queries/publicapi/search.sql`
- Create: `internal/publicapi/timeline.go`
- Create: `internal/publicapi/sources.go`
- Create: `internal/publicapi/search.go`
- Create: `internal/publicapi/publicapi_integration_test.go`

**Steps:**

1. Create indexed published read queries; use full-text/trigram aliases with bounded filters/pagination and PostGIS viewport query.
2. Test public isolation from draft/withdrawn/admin notes/raw payloads.
3. Return event/report/timeline revision ETags and `Last-Modified`; unchanged revalidation returns 304.
4. Search result states match source type and disclose highlights safely; measure query plans on representative volume.
5. Simulate ingestion outage and verify last published read models remain accessible with coverage degradation.

### Task 12: Event intelligence verification checkpoint

**Files:**

- Create: `docs/release-checklists/event-intelligence.md`
- Modify: `docs/06-data-model.md`
- Modify: `docs/08-reports-and-ai.md`

**Steps:**

1. Run golden correlation and report corpora; review every diff.
2. Run merge/split/publication concurrency and rollback tests under race detector.
3. Prove every published fact/report paragraph has a valid current or historical citation.
4. Disable all model credentials and rerun the full acceptance suite.
5. Measure table/index size and correlation/search latency against MVP data-volume fixture.
