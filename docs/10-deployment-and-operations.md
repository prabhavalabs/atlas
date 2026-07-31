# Deployment and operations

## Production topology

Atlas targets one Linux VPS with Docker Engine and the Compose plugin:

```text
Browser
  -> Cloudflare DNS/TLS/CDN
      -> outbound Cloudflare Tunnel
          -> Atlas app :8080 on atlas_edge
              -> PostgreSQL/PostGIS :5432 on internal atlas_data
```

The three core containers are `app`, `db`, and `cloudflared`. Atlas publishes no
host port, so it does not compete with an existing host reverse proxy. The
`atlas_data` network is internal; only the app can reach both networks. The Vite
assets, migrations, API, and operator commands live in one non-root distroless
application image.

## Resource profile

The practical foundation target is 4 shared vCPU, 8 GB RAM, and 80–160 GB SSD.
PostGIS receives the persistent `atlas_postgres` volume and 256 MB shared memory.
The app uses a read-only root filesystem, a 32 MB temporary filesystem, no Linux
capabilities, and bounded HTTP/database timeouts.

A local model or full global population dataset is not part of the core profile.
Those features require measured CPU, memory, storage, and failure isolation.

## Configuration and secrets

`.env.example` is the complete non-secret template. Production uses
`/opt/atlas/.env` with mode `0600`; it contains the database URL/password,
metrics token, public/API URLs, CORS origin, and cookie domain.
The Cloudflare tunnel token is a separate root-readable file at
`/opt/atlas/secrets/cloudflare_tunnel_token` and reaches cloudflared through
`TUNNEL_TOKEN_FILE`, so it does not appear in the process command line.

Required public endpoints are:

- `https://atlas.prabhavalabs.com` for the public application and `/admin`;
- `https://atlas-api.prabhavalabs.com` for `/api`, health, and protected metrics.

Both Cloudflare public hostnames route to `http://app:8080`. Browser credentials
work because both hosts are same-site and the CSRF cookie domain is
`prabhavalabs.com` so the readable, non-secret CSRF token is available to both
single-level application hosts. The opaque session cookie remains host-only on
`atlas-api.prabhavalabs.com`. Public CORS is an exact-origin allowlist.

## Operator commands

The image entrypoint accepts:

- `serve` — migrate, verify the web build, start HTTP, and shut down gracefully;
- `migrate` — apply embedded Goose migrations;
- `import-fixture --file <path>` — development/test data only;
- `create-admin --email <email> --name <name> --role <role>` — reads the password
  from standard input and stores only an Argon2id hash;
- `healthcheck` and `version` — container and release diagnostics.

Never import the synthetic fixture into a public production database.

## CI/CD

`.github/workflows/ci.yml` runs on pull requests and `main`:

1. generated SQL and OpenAPI-client drift checks;
2. Go race/unit tests, serialized PostGIS integration tests, vet, vulnerability
   scanning, and golangci-lint;
3. React unit/accessibility tests, lint, type checking, and production build;
4. a complete production-container build.

After successful `main` CI, `release.yml` builds one immutable
`ghcr.io/prabhavalabs/atlas:sha-<commit>` image, also advances `latest`, and uses a
dedicated SSH key to update only `/opt/atlas`. Compose waits for database and app
health. A failed rollout restores the previously running app image. It does not
restart, reconfigure, or remove unrelated Compose projects.

Repository deployment secrets are `ATLAS_VPS_HOST`, `ATLAS_VPS_USER`,
`ATLAS_VPS_SSH_KEY`, and `ATLAS_VPS_KNOWN_HOSTS`. The production environment can
add required reviewers without changing the workflow.

## Health and observability

- `/health/live` checks the Go process without dependencies.
- `/health/ready` verifies database reachability and schema compatibility.
- `/metrics` requires `Authorization: Bearer <ATLAS_METRICS_TOKEN>` and currently
  exposes the foundation `atlas_up` gauge. Request, job, source, and database-pool
  metrics are the next instrumentation increment.
- JSON lifecycle logs go to Docker stdout/stderr. Secret values and request bodies
  are never logged.
- cloudflared exposes its private metrics/ready endpoint only inside `atlas_edge`.

A free external uptime check can monitor public readiness. Cloudflare Tunnel is
outbound-only and maintains redundant edge connections; the core app remains
usable locally if public ingress is interrupted.

## Backup and restore

The database volume is persistent, but a volume is not a backup. Before real-source
beta, add a nightly custom-format `pg_dump`, encrypt it, upload to an operator-owned
S3-compatible target, retain 7 daily/4 weekly/6 monthly copies, and perform a
monthly isolated restore. Backup automation and a demonstrated restore remain a
beta release gate.

## Safe manual deployment

```sh
cd /opt/atlas
docker compose --env-file .env -f compose.yml config --quiet
docker compose --env-file .env -f compose.yml pull
docker compose --env-file .env -f compose.yml up -d --remove-orphans --wait --wait-timeout 180
docker compose --env-file .env -f compose.yml ps
```

Use an immutable `ATLAS_IMAGE` value. Record the previous image before changing
it. Migrations are forward-only; destructive contract migrations require a later
release after compatibility has been observed.

## Scaling path

1. Tune queries, indexes, retention, and cache headers.
2. Add job handlers or a second worker from the same image only when queue age
   demonstrates a need.
3. Move large raw/map artifacts to object storage.
4. Add an application replica and tunnel replica when availability requires it.
5. Extract population exposure only if it measurably interferes with the core.

Redis, a broker, Kubernetes, and public database ports are not baseline answers.
