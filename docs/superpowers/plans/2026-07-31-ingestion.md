# Ingestion and source reliability implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Ingest GDACS, USGS, NASA EONET, ReliefWeb, and one official CAP feed with replayable fixtures, isolated failures, visible freshness, and last-known-good behavior.

**Architecture:** Adapters fetch and decode source-specific representations; pure normalizers emit append-only observations. One transaction records run/observation versions and downstream jobs. Source health never mutates event severity.

**Tech stack:** Go HTTP/XML/JSON, pgx/sqlc, PostGIS, durable Postgres jobs, fixture replay/fuzz tests.

### Task 1: Finalize source registry and raw-artifact schema

**Files:**

- Create: `migrations/00005_source_registry.sql`
- Create: `queries/source/registry.sql`
- Create: `internal/source/model.go`
- Create: `internal/source/repository.go`
- Create: `internal/source/repository_integration_test.go`
- Create: `configs/sources.yaml`

**Steps:**

1. Write integration tests for unique stable key, configuration version, enabled/disabled state, terms/attribution fields, per-source freshness objective, and last-success versus last-change timestamps.
2. Add `source_runs` outcome enum values `success`, `not_modified`, `partial`, `failed`, and `disabled`; keep error class/message bounded.
3. Add optional `raw_artifacts` with content hash, storage reference, MIME, bytes, retrieval/retention times, and rights decision.
4. Load checked-in source definitions idempotently; runtime overrides are audited database values.
5. Validate every enabled source has terms URL, attribution, retention policy, contact User-Agent, and poll interval.

### Task 2: Build the bounded HTTP fetcher

**Files:**

- Create: `internal/source/httpfetch/client.go`
- Create: `internal/source/httpfetch/client_test.go`
- Create: `internal/source/httpfetch/security_test.go`

**Steps:**

1. Test connect/header/body timeouts, maximum bytes, gzip limits, accepted content types, ETag/Last-Modified, 304, 429 Retry-After, redirect policy, identifying User-Agent, and cancellation.
2. Reject redirects to non-HTTPS or unapproved hosts; resolve admin-submitted URLs through the SSRF policy before use.
3. Classify errors as transient, rate-limited, authentication, malformed, policy, or permanent.
4. Emit safe metrics without query strings or credentials.
5. Fuzz header and redirect parsing; run with race detector.

### Task 3: Implement the generic adapter runner

**Files:**

- Create: `internal/source/adapter.go`
- Create: `internal/source/runner.go`
- Create: `internal/source/runner_test.go`
- Create: `internal/observation/model.go`
- Create: `internal/observation/repository.go`
- Create: `internal/observation/repository_integration_test.go`

**Steps:**

1. Test success, unchanged, empty-success, partial decode, duplicate version, revised version, failure after prior success, and disabled source.
2. Make decoder output source candidates; make normalizer pure/no-network and validate time, units, geometry SRID/validity, status, and source link.
3. In one transaction record run, new versions, current pointer, and one idempotent `correlate_observation` job per changed observation.
4. Preserve the last successful snapshot and freshness after failure; increment error streak without creating empty observations.
5. Add advisory-lock schedule runner and per-source circuit breaker with manual reset/retry.

### Task 4: Implement USGS earthquake adapter

**Files:**

- Create: `internal/source/usgs/adapter.go`
- Create: `internal/source/usgs/normalize.go`
- Create: `internal/source/usgs/adapter_test.go`
- Create: `internal/source/usgs/fuzz_test.go`
- Create: `testdata/sources/usgs/`

**Steps:**

1. Add licensed/redacted fixtures for all-hour feed, detail, magnitude/depth revision, deletion, tsunami flag, missing magnitude, malformed coordinates, and empty feed.
2. Parse feed identity/update time; fetch detail only as asynchronous optional enrichment when it adds fields, never before core observation persistence.
3. Preserve magnitude value/type, depth/unit, event type/status, felt/significance/alert where supplied, and original links.
4. Test repeated import is idempotent and detail timeout leaves the feed observation usable with enrichment pending/unknown.
5. Add live schema monitor disabled in ordinary unit tests.

### Task 5: Implement GDACS adapter

**Files:**

- Create: `internal/source/gdacs/adapter.go`
- Create: `internal/source/gdacs/normalize.go`
- Create: `internal/source/gdacs/adapter_test.go`
- Create: `testdata/sources/gdacs/`

**Steps:**

1. Cover event-list GeoJSON for earthquake, flood, cyclone, volcano, drought/wildfire if present, green/orange/red alerts, polygon/track, revisions, missing estimates, and malformed item.
2. Preserve GDACS event ID/type, alert level/score, episode, from/to dates, geometry, population/impact estimates with model/approximate status, and source resources.
3. Do not convert missing alert/impact fields to zero. Quarantine invalid units/coordinates while retaining other valid items as a partial run.
4. Store mandatory attribution and expose GDACS automatic-estimate disclaimer through source metadata.
5. Verify source detail failure cannot change canonical severity/priority to low.

### Task 6: Implement NASA EONET adapter

**Files:**

- Create: `internal/source/eonet/adapter.go`
- Create: `internal/source/eonet/normalize.go`
- Create: `internal/source/eonet/adapter_test.go`
- Create: `testdata/sources/eonet/`

**Steps:**

1. Cover open/closed events, multiple categories/sources, point and polygon geometry history, magnitude units, duplicate source links, and deprecated/unknown categories.
2. Map EONET categories through versioned taxonomy; unknown categories remain ingestible and create review diagnostics rather than disappear.
3. Treat geometry date as validity time and preserve the geometry sequence.
4. Keep underlying curator/source URLs so EONET is not presented as the original authority when it is an aggregator.
5. Verify source closure is evidence, not unconditional canonical closure.

### Task 7: Implement ReliefWeb adapter

**Files:**

- Create: `internal/source/reliefweb/adapter.go`
- Create: `internal/source/reliefweb/normalize.go`
- Create: `internal/source/reliefweb/adapter_test.go`
- Create: `testdata/sources/reliefweb/`

**Steps:**

1. Register/require approved `appname`; test missing/invalid configuration before network use.
2. Ingest disaster metadata and recent reports using pagination/date cursors; preserve source, language, publication/original URLs, format, countries, disasters, and dates.
3. Store metadata and bounded excerpt only; do not persist full partner content unless the source registry permits it.
4. Canonicalize URLs and content hashes; link reports to candidate events by explicit ReliefWeb disaster relation first, geography/type/time second.
5. Test backfill cursor, duplicate page, report correction, removed item, and rate error recovery.

### Task 8: Implement CAP 1.2 reference adapter

**Files:**

- Create: `internal/source/cap/model.go`
- Create: `internal/source/cap/parser.go`
- Create: `internal/source/cap/normalize.go`
- Create: `internal/source/cap/parser_test.go`
- Create: `internal/source/cap/fuzz_test.go`
- Create: `testdata/sources/cap/`

**Steps:**

1. Test CAP Alert/Update/Cancel/Ack/Error, `Actual/Test/Exercise/System/Draft`, multilingual `info`, polygon/circle/geocode, references, effective/onset/expires, severity/urgency/certainty, and XML attacks.
2. Disable DTD/external entities and enforce XML depth/body limits.
3. Quarantine every non-`Actual` status from public correlation while retaining it for diagnostics. This must be an application and database constraint test.
4. Model Update/Cancel references explicitly and idempotently; preserve sender/identifier/sent tuple.
5. Configure NWS or another approved official CAP feed as the first source and follow its poll/User-Agent guidance.

### Task 9: Add compact geospatial reference imports

**Files:**

- Create: `migrations/00006_geo_reference.sql`
- Create: `internal/geo/importer.go`
- Create: `internal/geo/reverse.go`
- Create: `internal/geo/reverse_integration_test.go`
- Create: `cmd/atlas/import.go`
- Create: `configs/datasets.yaml`

**Steps:**

1. Document/download exact Natural Earth and GeoNames releases outside Git; verify checksum/license before import.
2. Import generalized admin polygons and selected populated places with GiST/trigram indexes, dataset version, attribution, and transformation metadata.
3. Test point-in-polygon, nearest city with distance/precision, border/antimeridian, disputed/unknown area, and deterministic results.
4. Return an approximate label/precision; never invent an affected area from a centroid.
5. Add attribution to public source metadata and `NOTICE`.

### Task 10: Build source health API/admin diagnostics

**Files:**

- Create: `internal/source/health.go`
- Create: `internal/publicapi/coverage.go`
- Create: `internal/adminapi/sources.go`
- Create: `internal/source/health_test.go`
- Modify: `api/openapi.yaml` (generated)

**Steps:**

1. Test healthy, stale, failing-with-last-good, never-successful, disabled, circuit-open, and partial states.
2. Public coverage reveals source class/region/hazard, last success, freshness category, and limitations without internal errors.
3. Admin diagnostics expose recent runs, bounded errors, next retry, and safe manual run/reset actions with audit reason.
4. Add source freshness alerts as metrics and review items, not external notification dependencies.
5. Verify a total simulated source outage leaves published list/detail available and visibly stale.

### Task 11: Ingestion verification checkpoint

**Files:**

- Create: `docs/release-checklists/ingestion.md`
- Modify: `docs/07-data-sources.md`

**Steps:**

1. Run fixtures/replay twice and prove zero duplicate observations/revisions/jobs.
2. Run fuzzers for a bounded CI corpus and race tests for runner/jobs.
3. Shadow-poll live sources without public publishing; compare counts/schema/freshness for at least 48 hours.
4. Simulate 429, 500, malformed data, timeout, and one-hour outage per source; record recovery.
5. Recheck terms/attribution/retention and ensure every enabled source entry is complete.
