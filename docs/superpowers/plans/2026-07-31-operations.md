# VPS operations and release implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Make a release reproducibly installable, observable, recoverable, secure, and affordable on one VPS with documented upgrades and failure playbooks.

**Architecture:** CI publishes a signed, SBOM-attached immutable image. Docker Compose runs Caddy, app, and PostGIS. Encrypted portable dumps/artifacts go to an operator-selected S3-compatible target. Health is visible in metrics and admin UI.

**Tech stack:** Docker/Compose, Caddy, GitHub Actions, PostgreSQL tools, restic or S3 CLI, Trivy/Grype, Syft, Cosign, shell smoke scripts.

### Task 1: Harden production images and Compose

**Files:**

- Modify: `Dockerfile`
- Modify: `deploy/compose.yaml`
- Modify: `deploy/Caddyfile`
- Create: `deploy/compose.ollama.yaml`
- Create: `deploy/compose.population.yaml`
- Create: `deploy/security.md`

**Steps:**

1. Pin base images by digest in release automation; use non-root app, minimal capabilities, read-only root filesystem/tmpfs, private DB network, explicit health/resource limits, and restart policies.
2. Keep only 80/443 public. Document SSH firewall, unattended security updates, time sync, and disk encryption provider considerations.
3. Test config with `docker compose config --quiet`; use container structure tests for user/files/health command/no package manager or build secrets.
4. Optional profiles cannot alter core app readiness and remain off by default.
5. Verify clean host brings up exactly three long-running core containers.

### Task 2: Implement migration preflight, upgrade, and rollback

**Files:**

- Create: `deploy/scripts/preflight.sh`
- Create: `deploy/scripts/upgrade.sh`
- Create: `deploy/scripts/rollback.sh`
- Create: `deploy/scripts/migration-check.sh`
- Create: `docs/runbooks/upgrade-rollback.md`

**Steps:**

1. Preflight checks Docker/Compose version, architecture, free disk/RAM, env completeness, backup recency, image signature, and database compatibility.
2. Upgrade pulls explicit version, creates pre-upgrade dump, runs migration one-shot, starts app, waits readiness, and runs smoke.
3. Rollback supports application rollback when schema remains backward compatible; schema restore requires explicit documented destructive recovery.
4. CI migrates empty and previous release databases and starts both old/new app compatibility windows.
5. Inject migration failure and prove old version/data remain recoverable.

### Task 3: Implement encrypted backup and restore

**Files:**

- Create: `deploy/scripts/backup.sh`
- Create: `deploy/scripts/restore.sh`
- Create: `deploy/scripts/prune-backups.sh`
- Create: `deploy/systemd/atlas-backup.service`
- Create: `deploy/systemd/atlas-backup.timer`
- Create: `docs/runbooks/backup-restore.md`

**Steps:**

1. Test scripts against an isolated S3-compatible test target: portable `pg_dump --format=custom`, application-data archive, checksum, client-side encryption, upload, and retention.
2. Never place encryption/backup credentials in command arguments/logs; validate target prefix to prevent broad deletion.
3. Restore into a new database/volume, run migrations/readiness, and verify published event/report/audit hashes.
4. Record backup/restore result, bytes, duration, restorable timestamp, and checksum in health metadata.
5. Automate 7 daily/4 weekly/6 monthly retention with dry-run and exact-prefix safety; do not delete on failed backup.

### Task 4: Add observability and actionable alerting

**Files:**

- Create: `deploy/monitoring/prometheus-example.yml`
- Create: `deploy/monitoring/alerts.yml`
- Create: `docs/runbooks/observability.md`
- Modify: `internal/platform/observability/metrics.go`

**Steps:**

1. Define alerts for public readiness, source freshness by objective, job oldest age/dead jobs, database connections/disk estimate, backup/restore age, and repeated report failures.
2. Keep labels bounded; test no event title/URL/error body creates metric cardinality or leaks data.
3. Provide optional Prometheus configuration but do not add it to core Compose. Document external HTTP uptime monitor alternative.
4. Add `/api/v1/admin/health` aggregation tested against thresholds and missing telemetry.
5. Run a controlled source failure and confirm public stale state, admin action, metric, and recovery all align.

### Task 5: Implement retention and capacity safeguards

**Files:**

- Create: `internal/platform/retention/service.go`
- Create: `internal/platform/retention/service_test.go`
- Create: `queries/platform/retention.sql`
- Create: `docs/runbooks/capacity.md`

**Steps:**

1. Test raw-artifact expiry by source policy, fetch/job log retention, publication/audit indefinite preservation, batch limits, and interruption resume.
2. Run retention through durable jobs with dry-run counts and audit/metrics.
3. Add admin table/index/artifact/backup size report and thresholds at 60/75/90% disk.
4. Use exact IDs/retention timestamps, never unvalidated paths/globs for deletion.
5. Load representative one-year fixture volume and measure growth/index bloat/retention time.

### Task 6: Build CI supply-chain and release pipeline

**Files:**

- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `.github/dependabot.yml`
- Create: `.goreleaser.yaml`
- Create: `SECURITY.md`
- Create: `CONTRIBUTING.md`

**Steps:**

1. CI runs format/static/race/unit/integration/frontend/axe/e2e, OpenAPI/sqlc drift, migration, docs links, secrets, licenses, container scan, and Compose smoke.
2. Release on signed version tag builds amd64/arm64, emits checksums/SBOM/provenance, signs image/artifacts, and never publishes `latest` as deployment input.
3. Pin actions by commit and give minimal workflow permissions; environment secrets only in release job.
4. Add grouped automated dependency PRs and documented supported-version policy.
5. Verify release from a clean tag and signature/SBOM before VPS preflight accepts it.

### Task 7: Security and abuse verification

**Files:**

- Create: `test/security/hostile_server.go`
- Create: `docs/release-checklists/security.md`
- Modify: `docs/09-security-and-trust.md`

**Steps:**

1. Test feed decompression bomb/XXE/deep JSON, SSRF private/link-local/redirect/DNS-rebind defenses, stored/reflected XSS, SQL/filter abuse, oversized geometry, and auth/CSRF/session/rate controls.
2. Scan dependencies/images/IaC/secrets; triage every high/critical issue with fix or recorded non-applicability.
3. Load-test anonymous APIs and login/admin mutations; verify bounded CPU/memory/database connections and fair public service.
4. Publish security contact/vulnerability process and define secret-rotation/session-revocation procedure.
5. Review optional model prompt injection boundary and prove it has no tools/network/publish permission.

### Task 8: Execute failure and disaster-recovery drills

**Files:**

- Create: `deploy/scripts/drill.sh`
- Create: `docs/runbooks/source-outage.md`
- Create: `docs/runbooks/database-full.md`
- Create: `docs/runbooks/vps-loss.md`
- Create: `docs/runbooks/incorrect-publication.md`
- Create: `docs/release-checklists/disaster-recovery.md`

**Steps:**

1. Drill one/all source outage, map outage, model outage, database restart, process kill with leased job, backup target outage, job poison, disk-near-full, and incorrect publication.
2. Record expected detection, degraded public behavior, operator action, recovery, and post-check for each.
3. Restore latest backup onto a clean host and measure RPO/RTO; target RPO 24h and RTO 4h.
4. Verify public last-known-good/static low-zoom behavior during upstream failures.
5. Turn unexpected drill behavior into tests/issues before release.

### Task 9: Benchmark the core VPS and set limits

**Files:**

- Create: `test/load/k6-public.js`
- Create: `test/load/ingestion-burst.go`
- Create: `docs/benchmarks/core-vps.md`

**Steps:**

1. Seed representative events/observations/timeline/articles and measure indexed public list/detail/search/viewport plus admin workspace.
2. Test normal traffic, 10x disaster spike, ingestion burst, job outage recovery, and concurrent publication.
3. Record p50/p95/p99, error rate, CPU, memory, DB pool, disk I/O, payload/bundle size on 4-vCPU/8-GB equivalent.
4. Tune queries/indexes/cache before raising resources; attach `EXPLAIN (ANALYZE, BUFFERS)` for slow queries.
5. Set container/query/body/concurrency limits and regression thresholds in CI/nightly tests.

### Task 10: Prepare and execute public beta release

**Files:**

- Create: `docs/release-checklists/beta.md`
- Create: `docs/self-hosting.md`
- Create: `docs/operator-guide.md`
- Create: `CHANGELOG.md`

**Steps:**

1. Recheck product name, license, every source/map/model/data term and attribution, privacy/correction/accessibility/security pages.
2. Complete all phase checklists, real-source shadow observation, golden evaluation, security/load/restore/rollback drills.
3. Install signed candidate on a clean VPS using only public docs and time the process; fix every undocumented step.
4. Publish beta with monitoring disclaimer, known coverage gaps, cost baseline, reproducible release notes, and feedback channels.
5. Observe for at least one full source cadence cycle before enabling optional model renderer or additional sources.

### Task 11: Post-release maintenance cadence

**Files:**

- Create: `docs/maintenance.md`
- Create: `.github/ISSUE_TEMPLATE/source-problem.yml`
- Create: `.github/ISSUE_TEMPLATE/incorrect-information.yml`

**Steps:**

1. Document daily health review, weekly dependency/source triage, monthly update/restore, quarterly capacity/cost/terms review, and annual DR/key exercise.
2. Track source adapter owner/last-verified date and display overdue maintenance.
3. Publish actual monthly resource/cost report and compare with `docs/11-cost-model.md`.
4. Accept new sources/features only with scope, terms, fixtures, failure behavior, and complexity-budget review.
5. Revisit modular extraction only when measurements satisfy ADR criteria.
