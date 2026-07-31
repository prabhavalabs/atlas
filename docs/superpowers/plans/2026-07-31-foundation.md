# Repository and application foundation implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Create a tested, deployable walking skeleton that moves one fixture observation through PostGIS to an anonymous public event page and records one authenticated admin action.

**Architecture:** One Go command assembles modules and serves generated OpenAPI plus embedded Vite assets. Postgres owns durable state. The first vertical slice establishes conventions that later source and editorial work must follow.

**Tech stack:** Go 1.26, chi, pgx/sqlc, Goose, OpenAPI 3.1, PostgreSQL/PostGIS, React/Vite/TypeScript, TanStack Router/Query, Zustand, shadcn/ui, Docker Compose, and Cloudflare Tunnel.

### Task 1: Bootstrap license, toolchain, and repository checks

**Files:**

- Create: `LICENSE`
- Create: `NOTICE`
- Create: `README.md`
- Create: `.gitignore`
- Create: `.editorconfig`
- Create: `.env.example`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.github/workflows/ci.yml`

**Steps:**

1. Preserve the accepted Apache-2.0 license text and record data/component notices separately in `NOTICE`.
2. Initialize module `github.com/prabhavalabs/atlas` with Go 1.26 and set `toolchain` to the current 1.26 patch.
3. Add Make targets `fmt`, `lint`, `test`, `test-integration`, `generate`, `build`, and `verify`.
4. Add a CI smoke test that intentionally fails until `go test ./...` and frontend scripts exist, then complete them in later tasks.
5. Run `go mod tidy` and `go test ./...`; commit only after both pass.

### Task 2: Add validated configuration and application assembly

**Files:**

- Create: `cmd/atlas/main.go`
- Create: `internal/platform/config/config.go`
- Create: `internal/platform/config/config_test.go`
- Create: `internal/platform/app/app.go`
- Create: `internal/platform/buildinfo/buildinfo.go`

**Steps:**

1. Write failing table tests for missing database URL, invalid public URL, invalid log level, and a minimal valid test configuration.
2. Implement a typed `Config` with explicit defaults and `Validate() error`; do not read environment variables outside the config package.
3. Implement subcommands `serve`, `worker`, `migrate`, `admin`, and `import` with help text. Unimplemented subcommands return a clear non-zero error until their task lands.
4. Inject build version, commit, and build time through linker flags and expose a plain Go `BuildInfo` value.
5. Run `go test ./internal/platform/config ./internal/platform/app ./cmd/atlas` and verify errors contain the setting name but never secret values.

### Task 3: Establish PostgreSQL/PostGIS migrations and typed queries

**Files:**

- Create: `migrations/00001_platform.sql`
- Create: `internal/platform/database/database.go`
- Create: `internal/platform/database/database_test.go`
- Create: `queries/platform/health.sql`
- Create: `sqlc.yaml`
- Create: `internal/platform/store/` (generated)
- Create: `internal/testsupport/postgres.go`

**Steps:**

1. Write an integration test that creates an empty PostGIS database, runs migrations, checks `postgis_full_version()`, and runs a typed `SELECT 1` query.
2. Migration `00001` enables `postgis`, `pg_trgm`, and `citext`; creates `schema_metadata`; and records migration/application compatibility metadata.
3. Configure sqlc with pgx/v5 and strict query checks. Generated code is checked in and `make generate` fails on drift.
4. Implement database pool timeouts, health check, statement timeout, and shutdown.
5. Add CI PostGIS service and run `make test-integration` from an empty database.

### Task 4: Implement HTTP, errors, health, and OpenAPI

**Files:**

- Create: `internal/platform/httpserver/server.go`
- Create: `internal/platform/httpserver/middleware.go`
- Create: `internal/platform/httpserver/errors.go`
- Create: `internal/platform/httpserver/server_test.go`
- Create: `internal/publicapi/meta.go`
- Generate: `api/openapi.yaml`

**Steps:**

1. Write failing handler tests for `/health/live`, `/health/ready`, `/api/v1/meta`, unknown route JSON problem response, request ID, method rejection, and body limit.
2. Assemble chi handlers with recovery, request ID, security headers, timeouts, explicit CORS, and a checked-in OpenAPI contract.
3. Make liveness dependency-free; readiness checks database and schema compatibility.
4. Return RFC 9457-style problem details with stable error codes and no internal error text.
5. Generate OpenAPI, check it in, and add `make generate-check` that fails when regeneration changes files.

### Task 5: Add structured logging and metrics

**Files:**

- Create: `internal/platform/observability/log.go`
- Create: `internal/platform/observability/metrics.go`
- Create: `internal/platform/observability/observability_test.go`

**Steps:**

1. Test JSON log fields for request ID, job ID, source key, and event ID; test redaction of password, token, authorization, cookie, and raw body fields.
2. Add Prometheus counters/histograms for requests, jobs, source runs, and database pool state with bounded labels.
3. Serve `/metrics` only on a configured private/admin listener or require an internal token; do not expose source URLs containing credentials.
4. Verify a 500 response is correlated with one structured error record and no duplicate stack spam.

### Task 6: Create the jobs table and one idempotent worker

**Files:**

- Create: `migrations/00002_jobs.sql`
- Create: `queries/jobs/jobs.sql`
- Create: `internal/jobs/model.go`
- Create: `internal/jobs/repository.go`
- Create: `internal/jobs/worker.go`
- Create: `internal/jobs/worker_test.go`
- Create: `internal/jobs/repository_integration_test.go`

**Steps:**

1. Write integration tests for enqueue idempotency, concurrent `SKIP LOCKED` claims, lease expiry/reclaim, successful completion, exponential retry, and poison-job terminal state.
2. Create `jobs` with a unique `(type, idempotency_key)`, state constraint, attempt/availability/lease fields, JSON payload, and diagnostic error fields.
3. Implement a typed handler registry. Unknown job types move to a visible terminal failure rather than retry forever.
4. Use an injectable clock/random jitter source for deterministic tests.
5. Kill a worker after claim in an integration test; verify another worker resumes after lease expiry.

### Task 7: Create minimal source, observation, and event vertical slice

**Files:**

- Create: `migrations/00003_walking_skeleton.sql`
- Create: `queries/source/sources.sql`
- Create: `queries/observation/observations.sql`
- Create: `queries/event/events.sql`
- Create: `internal/source/fixture/adapter.go`
- Create: `internal/observation/service.go`
- Create: `internal/event/service.go`
- Create: `internal/event/service_integration_test.go`
- Create: `testdata/sources/fixture/minimal-event.json`
- Create: `internal/publicapi/events.go`
- Create: `internal/publicapi/events_test.go`

**Steps:**

1. Write an end-to-end integration test: import the fixture, create versioned observation, create canonical draft event, publish it, list anonymously, and retrieve by slug.
2. Add only the minimal tables/columns from `docs/06-data-model.md`; use UUIDs, UTC timestamps, geometry SRID constraints, and revision foreign keys.
3. Make import idempotent using source key/external ID/content hash.
4. Return public event DTOs rather than database rows. Include ETag, last modified, data generated time, and coverage state.
5. Re-import unchanged fixture and assert no new revision/timeline noise.

### Task 8: Add local admin identity and audit skeleton

**Files:**

- Create: `migrations/00004_identity_audit.sql`
- Create: `queries/identity/identity.sql`
- Create: `queries/audit/audit.sql`
- Create: `internal/identity/password.go`
- Create: `internal/identity/session.go`
- Create: `internal/identity/session_test.go`
- Create: `internal/audit/service.go`
- Create: `internal/adminapi/session.go`
- Create: `internal/adminapi/events.go`
- Create: `internal/adminapi/adminapi_test.go`

**Steps:**

1. Test Argon2id hash/verify, uniform invalid-login response, session rotation/revocation/expiry, CSRF failure, secure cookie flags, role denial, and rate limiting.
2. Implement interactive `atlas admin create` reading password without echo; never accept password as a command-line argument.
3. Implement `POST /api/v1/admin/session`, `DELETE`, and one revision-checked event title mutation.
4. In the same transaction as the mutation, append an audit record with actor, reason, target, before/after hashes, and request ID.
5. Assert anonymous public reads still work and every `/admin` mutation fails without session and CSRF token.

### Task 9: Bootstrap the React/Vite workspace and generated client

**Files:**

- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/src/main.tsx`
- Create: `web/src/router.tsx`
- Create: `web/src/app/providers.tsx`
- Create: `web/src/routes/public.tsx`
- Create: `web/src/routes/admin.tsx`
- Create: `web/src/lib/api/` (generated)
- Create: `web/src/styles/globals.css`
- Create: `web/src/test/setup.ts`

**Steps:**

1. Pin React/Vite/TypeScript, TanStack Router/Query, Zustand, Tailwind, shadcn/Radix dependencies, Vitest, Testing Library, MSW, axe, and Playwright.
2. Generate the API client from `api/openapi.yaml`; fail CI on drift.
3. Configure QueryClient defaults: bounded retry by error class, no mutation retry by default, and visible stale data.
4. Create code-split public and `/admin` route shells; admin shell requests session state, public shell never does.
5. Write tests for anonymous public route, unauthenticated admin login route, query error boundary, and keyboard skip link.

### Task 10: Build the walking-skeleton public/admin screens

**Files:**

- Create: `web/src/features/events/event-list.tsx`
- Create: `web/src/features/events/event-page.tsx`
- Create: `web/src/features/auth/login-form.tsx`
- Create: `web/src/features/admin/event-title-form.tsx`
- Create: `web/src/routes/__tests__/walking-skeleton.test.tsx`
- Create: `web/e2e/walking-skeleton.spec.ts`

**Steps:**

1. Write MSW-backed tests for list/detail loading, empty, stale, failed, and successful states; add axe assertions.
2. Render the fixture event in a semantic list and event narrative with source/freshness labels.
3. Implement admin login and revision-checked title mutation with validation/conflict display.
4. Keep server data in TanStack Query; use Zustand only for one local preference to establish the boundary.
5. Playwright: anonymous visitor views event; admin signs in, edits title with reason, signs out; visitor sees updated published revision only after publication policy permits it.

### Task 11: Package the three-container deployment

**Files:**

- Create: `Dockerfile`
- Create: `deploy/compose.yml`
- Create: `deploy/secrets/.gitkeep`
- Create: `deploy/.env.example`
- Create: `deploy/scripts/smoke.sh`
- Create: `docs/self-hosting.md`

**Steps:**

1. Build frontend and Go binary in pinned build stages; final non-root image contains CA certificates, migrations, and frontend assets only.
2. Add cloudflared, app, and PostGIS services on isolated edge/data networks with health checks and an explicit persistent volume.
3. Run migration as an explicit one-shot command before app startup; do not auto-run destructive migrations in every replica.
4. Write smoke script using public meta/list/detail and admin login CSRF checks.
5. From a clean Docker state, run `docker compose up -d`, import fixture, smoke test, restart all containers, and verify data persists.

### Task 12: Foundation verification and checkpoint

**Files:**

- Modify: `docs/README.md`
- Create: `docs/release-checklists/foundation.md`

**Steps:**

1. Run all master phase commands, race tests, Compose smoke, OpenAPI drift, axe, and Playwright.
2. Capture database size, idle memory, startup time, and fixture ingest latency as baseline in the checklist.
3. Verify `rg -i` finds no predecessor name, private repository identifier, secret, or internal endpoint in public files/history.
4. Review dependency licenses and generate initial SBOM/NOTICE inventory.
5. Commit the walking skeleton only when a clean clone can reproduce it from README/self-hosting docs.
