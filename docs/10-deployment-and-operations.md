# Deployment and operations

## Production topology

Atlas targets one Linux VPS with Docker Engine, the Compose plugin, and a shared
host Nginx:

```text
Browser
  -> Cloudflare proxied A records and edge TLS
      -> host Nginx :80/:443 with Let's Encrypt origin TLS
          -> 127.0.0.1:8081
              -> Atlas app on atlas_edge
                  -> PostgreSQL/PostGIS :5432 on internal atlas_data
```

The two Atlas containers are `app` and `db`. Compose publishes the application
only on `127.0.0.1:8081`; remote hosts cannot reach that port. Atlas reuses the
VPS's existing Nginx listener and does not start a second reverse proxy. The
`atlas_data` network is internal, and PostgreSQL has no host port.

The Vite assets, migrations, API, and operator commands live in one non-root
distroless application image.

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
metrics token, public/API URLs, CORS origin, and cookie domain. No ingress token
is required.

Required public endpoints are:

- `https://atlas.prabhavalabs.com` for the public application and `/admin`;
- `https://atlas-api.prabhavalabs.com` for `/api`, health, and protected metrics.

Both hostnames reach the same modular-monolith process through Nginx. Browser
credentials work because the hosts are same-site and the CSRF cookie domain is
`prabhavalabs.com`. The opaque session cookie remains host-only on
`atlas-api.prabhavalabs.com`. Public CORS is an exact-origin allowlist.

## One-time Nginx provisioning

The repository carries a bootstrap HTTP virtual host, the final HTTPS virtual
host, Cloudflare's origin allowlist, and the renewal hook in `deploy/nginx`.
Provision these as root without replacing any other Nginx site:

```sh
install -d -m 0755 /var/www/atlas-acme /etc/nginx/snippets
install -m 0644 deploy/nginx/cloudflare-origin-allow.conf \
  /etc/nginx/snippets/atlas-cloudflare-origin-allow.conf
install -m 0644 deploy/nginx/atlas-bootstrap.conf \
  /etc/nginx/sites-available/atlas
ln -s /etc/nginx/sites-available/atlas /etc/nginx/sites-enabled/atlas
nginx -t
systemctl reload nginx
```

If the symlink already exists, do not recreate it. Validate the bootstrap route
locally before changing DNS:

```sh
curl --fail --header 'Host: atlas.prabhavalabs.com' http://127.0.0.1/health/ready
curl --fail --header 'Host: atlas-api.prabhavalabs.com' http://127.0.0.1/health/ready
```

## DNS and origin TLS

Create two Cloudflare DNS records with TTL Auto. Keep the proxy **DNS only**
during certificate issuance:

| Type | Name | IPv4 address | Proxy status |
|---|---|---|---|
| A | `atlas` | VPS public IPv4 | DNS only initially |
| A | `atlas-api` | VPS public IPv4 | DNS only initially |

The bootstrap virtual host exposes only the ACME challenge path; all other remote
HTTP requests are denied. Keeping the records DNS-only lets Let's Encrypt reach
that challenge without depending on the zone's current redirect or encryption
settings.

After both names resolve and the ACME challenge path is reachable, issue one
certificate covering both hosts, install the final virtual host, and test
renewal:

```sh
apt-get update
apt-get install -y certbot
certbot certonly --webroot --webroot-path /var/www/atlas-acme \
  --cert-name atlas.prabhavalabs.com \
  --domain atlas.prabhavalabs.com \
  --domain atlas-api.prabhavalabs.com
install -m 0644 deploy/nginx/atlas.conf /etc/nginx/sites-available/atlas
install -m 0755 deploy/nginx/reload-nginx \
  /etc/letsencrypt/renewal-hooks/deploy/atlas-reload-nginx
nginx -t
systemctl reload nginx
certbot renew --dry-run
```

After the origin certificate is active, enable the orange-cloud proxy on both A
records and set the Cloudflare zone encryption mode to **Full (strict)**. This
keeps browser-to-edge and edge-to-origin traffic encrypted. Never weaken the
whole zone to Flexible for Atlas. Before enabling the proxy, compare
`deploy/nginx/cloudflare-origin-allow.conf` with Cloudflare's canonical
[`ips-v4`](https://www.cloudflare.com/ips-v4/) and
[`ips-v6`](https://www.cloudflare.com/ips-v6/) lists.

### Migrating an existing tunnel deployment

Keep the connector and its DNS records active while staging the loopback port
and bootstrap virtual host. Do not run `docker compose --remove-orphans` yet.
After the two Nginx host-header smoke checks pass:

1. replace both tunnel CNAMEs with DNS-only A records;
2. issue the origin certificate and install the final Nginx virtual host;
3. enable the proxy on both A records and verify public HTTPS;
4. deploy the current Compose file with `--remove-orphans` to stop the connector;
5. revoke and remove the unused tunnel only after an observation window.

This sequence keeps the old route recoverable until the new route is proven.

## Forwarded-address trust

The login limiter uses `CF-Connecting-IP`. The Nginx virtual host is therefore a
security boundary: it accepts only Cloudflare source networks and overwrites the
forwarding headers before proxying. Do not remove the origin allowlist or publish
port `8081` on a non-loopback address. Review the source ranges quarterly and
whenever Cloudflare announces an address change.

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
health. Before the first tunnel-free release, the workflow requires both local
HTTPS hostnames to pass certificate and readiness checks; until then, it exits
without changing the current deployment. A failed rollout restores the previous
Compose definition and application image. It does not restart, reconfigure, or
remove unrelated Compose projects or the shared Nginx.

Repository deployment secrets are `ATLAS_VPS_HOST`, `ATLAS_VPS_USER`,
`ATLAS_VPS_SSH_KEY`, and `ATLAS_VPS_KNOWN_HOSTS`. The production environment can
add required reviewers without changing the workflow.

## Health and observability

- `/health/live` checks the Go process without dependencies.
- `/health/ready` verifies database reachability and schema compatibility.
- `/metrics` requires `Authorization: Bearer <ATLAS_METRICS_TOKEN>` and currently
  exposes the foundation `atlas_up` gauge.
- JSON lifecycle logs go to Docker stdout/stderr. Secret values and request bodies
  are never logged.
- Nginx access and error logs cover public ingress; Cloudflare provides edge
  request identifiers in `CF-Ray`.

A free external uptime check can monitor public readiness. The core app remains
usable through the loopback route if Cloudflare is interrupted.

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
curl --fail http://127.0.0.1:8081/health/ready
```

Use an immutable `ATLAS_IMAGE` value. Record the previous image before changing
it. Migrations are forward-only; destructive contract migrations require a later
release after compatibility has been observed.

## Scaling path

1. Tune queries, indexes, retention, and cache headers.
2. Add job handlers or a second worker from the same image only when queue age
   demonstrates a need.
3. Move large raw/map artifacts to object storage.
4. Add an application replica and update the Nginx upstream when availability
   requires it.
5. Extract population exposure only if it measurably interferes with the core.

Redis, a broker, Kubernetes, and public database ports are not baseline answers.
