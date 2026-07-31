# Testing and quality strategy

## Foundation quality gates

The current walking skeleton runs Go unit tests with the race detector, real
PostGIS integration tests, `go vet`, `govulncheck`, golangci-lint, generated SQL
and OpenAPI-client drift checks, React flow tests, an axe accessibility check,
ESLint, TypeScript checking, a production frontend build, and a production image
build in CI. Local Compose/image smoke tests cover migrations, readiness, public
API responses, and SPA delivery.

The remaining sections define the test strategy as ingestion, correlation,
publication, backup/restore, and report generation are added. A listed target is
not a claim that the corresponding subsystem already exists.

## Test pyramid by risk

### Go unit and property tests

- source parsing/normalization from immutable fixtures;
- units, timezones, geometry validity, hashes, canonical URLs;
- correlation feature calculation and thresholds;
- lifecycle, severity, priority, hysteresis, and material-change rules;
- report claim/citation validation;
- job idempotency and retry classification.

Fuzz targets cover XML/JSON/feed decoders, geometry input, URL parsing, and report output validation.

### Database integration tests

Use a real PostGIS test container for migrations, generated queries, spatial correlation, full-text search, leases, constraints, and transactions. CI tests:

- migrate from empty database;
- migrate from the previous released schema;
- application enum/config values against database constraints;
- downgrade/rollback procedure where supported;
- simultaneous job claims and scheduler leadership;
- event merge/split and report/publication atomicity.

Mocks do not replace these tests.

### Adapter contract/replay tests

Each source has checked-in redacted fixtures for normal, empty, update, retraction, malformed, rate-limited, unauthorized, oversized, future-dated, test/exercise, and partial responses. A nightly non-mutating contract monitor checks live schemas and records drift without making production correctness depend on the monitor.

### Frontend tests

- Vitest and Testing Library for components and flows.
- MSW generated from OpenAPI examples for network states.
- axe checks on key routes/components.
- Playwright for public overview/event and admin review/publish/correction journeys.
- Visual regression at desktop/mobile, light/dark, stale/degraded/empty states.
- Contract generation must have no diff after backend changes.

### End-to-end system tests

Start the production Compose profile, ingest fixtures, review/publish an event, validate public cache behavior, create a correction, back up, destroy the test database, restore, and verify the published revision/audit history.

## Golden event corpus

Maintain licensed/redacted examples for:

- earthquake with repeated magnitude/depth updates and deletion;
- cyclone with moving track and forecast-cone revisions;
- flood with uncertain geometry and humanitarian reports;
- wildfire detections that must not each become an event;
- CAP alert update/cancel/test/exercise messages;
- two nearby but separate disasters;
- one event represented by several sources/languages;
- conflicting casualty figures;
- complete source outage and recovery;
- report correction after publication.

Every correlation/scoring/report change runs against this corpus and reports diffs for review.

## Performance budgets

Test on hardware equivalent to the core VPS:

- cached public event list/detail p95 under 500 ms;
- uncached indexed list/search p95 under 1 second at expected dataset size;
- event page JSON under 500 KB before compression, with paged source history;
- public initial JS budget set during design foundation and enforced in CI;
- ingestion remains current during a 10x fixture burst;
- job queue recovers after a one-hour simulated upstream outage;
- memory remains within container limits during geometry/report workloads.

Exact frontend bundle budget is chosen after the map/code-splitting spike; map and admin bundles must load lazily.

## Reliability tests

Inject: timeout, DNS failure, 429/500, malformed body, clock skew, duplicate delivery, process kill during transaction/job, database restart, disk-near-full, stale lease, model invalid JSON, map-provider failure, and backup-target outage. Verify last-known-good public data, visible degradation, bounded retry, and recovery.

## Quality gates

A pull request must pass formatting, lint/static analysis, unit/integration/frontend tests, OpenAPI drift, migration checks, license/secret/dependency scans, accessibility checks on affected routes, and docs link checks. Main branch produces a reproducible signed image; releases require Compose smoke, restore, and rollback evidence.

## Coverage policy

Line coverage is reported but risk coverage is the gate. Source decoders, correlation, lifecycle, publication, auth, migrations, and report validators require branch/edge-case tests. A high global percentage cannot compensate for an untested state transition.
