# Deployment and operations

## Production topology

The default target is one Linux VPS with Docker Engine and the Compose plugin:

```text
Internet
  -> Caddy :80/:443
      -> Go application :8080 (private Docker network)
          -> PostgreSQL/PostGIS :5432 (private Docker network)

Persistent volumes
  - postgres-data
  - application-data (raw artifacts, optional low-zoom PMTiles)
  - caddy-data
```

Only ports 80, 443, and restricted SSH are exposed. PostgreSQL is never public. The application image contains the Go binary, migrations, source defaults, public/admin frontend assets, and build metadata.

## Resource profiles

### Core MVP

- 4 shared vCPU
- 8 GB RAM
- 80–160 GB SSD
- 1–2 GB swap for crash protection, not normal operation

PostgreSQL receives a conservative memory budget; Go and Caddy remain small. Raw artifact retention and database growth are measured from day one.

### Core plus local model or full population service

- 8 vCPU
- 16–32 GB RAM depending on model
- 200+ GB SSD for database/model/data

These are optional profiles. A local model must not contend with ingestion/database memory on the core profile. The full population dataset is benchmarked before it is advertised as a supported profile.

## Compose services

- `proxy`: pinned Caddy image, read-only config, persistent certificate storage.
- `app`: non-root, read-only root filesystem where possible, health/readiness endpoints, explicit memory/CPU limits.
- `db`: pinned PostGIS image, checksummed migrations, local-only network, persistent volume.

Profiles may add `backup`, `ollama`, or `population`, but core production health never depends on them.

## Configuration

Configuration is environment-driven with a checked-in `.env.example` containing no secrets. Startup validates every setting and exits with actionable errors. Secrets include database password, session key, optional source keys, optional model key, and backup credentials.

Runtime-tunable source enablement and polling cadence live in the database with audit history. Security-sensitive changes require restart/environment configuration.

## Deployment workflow

1. CI builds and tests backend/frontend.
2. Generate OpenAPI/client and fail on uncommitted drift.
3. Build multi-architecture container image with SBOM and immutable version tag.
4. Scan and sign the image.
5. VPS pulls the explicit version, never `latest`.
6. Run backward-compatible migrations as a one-shot command.
7. Start the new app and wait for readiness.
8. Run public/admin/source smoke tests.
9. Keep the previous image and documented rollback command.

Migrations follow expand/migrate/contract discipline. A release cannot depend on a schema value that its included migration has not created. Destructive schema cleanup is a later release after compatibility has been observed.

## Health and observability

- `/health/live`: process is alive; no dependency checks.
- `/health/ready`: database reachable, migrations current, application initialized.
- `/metrics`: request latency/status, job queue/age, source freshness/errors, report generation, database pool.
- Structured JSON logs include request/job/source/event IDs and redact secrets/content bodies.
- Admin system-health screen translates operational metrics into actions.

Alerting can initially use a free external HTTP uptime monitor plus a daily local health digest. Inbound monitoring must not be necessary for correctness.

## Backup and restore

- Nightly `pg_dump` in custom format, compressed and encrypted before off-site upload.
- Daily application-data snapshot for retained raw artifacts and local map fallback.
- Keep 7 daily, 4 weekly, and 6 monthly backups initially.
- Validate checksum after upload.
- Perform a documented restore into an isolated database every month.
- Record restore duration and the latest restorable point in admin health.

[Backblaze B2](https://www.backblaze.com/cloud-storage/pricing) is a low-cost S3-compatible example, currently including the first 10 GB free according to its transaction pricing page. Any S3-compatible provider works. Restic is an acceptable encrypted file-backup client; database dumps remain the portable recovery artifact.

## Updates and maintenance

- Weekly automated dependency update PRs, grouped by risk.
- Monthly production update window and restore drill.
- Supported Go and PostgreSQL/PostGIS release policy documented.
- Source adapters have owner and “last verified” date; stale adapters become visible maintenance work.
- Quarterly retention/capacity review.
- Annual key rotation and disaster-recovery exercise.

## Failure playbooks

Runbooks cover: one source stale, all sources stale, database full, migration failure, job backlog, corrupt raw payload, map provider unavailable, model provider unavailable, compromised admin session, incorrect public report, and VPS loss. The expected degraded behavior is defined in tests, not only in prose.

## Scaling path

1. Tune indexes/queries and cache headers.
2. Add a read replica only if database metrics justify it.
3. Run a second worker from the same application image for job throughput.
4. Move large artifacts/map archives to object storage/CDN.
5. Extract population exposure only if its resources interfere with core monitoring.
6. Move to multiple application replicas with shared Postgres and advisory-lock scheduling.

The application remains useful on the single-VPS topology throughout.
