# ADR 0008: Use the shared host Nginx for ingress

- Status: Accepted
- Date: 2026-08-01

## Context

Atlas needs two public HTTPS hostnames on a VPS that already has a host-level
Nginx listener. The first deployment used an outbound tunnel to avoid port
collisions, but that added a connector, a privileged token, and an additional
availability dependency. Replacing the existing listener with Caddy would risk
unrelated applications on the shared host.

## Decision

Publish the Atlas application only on `127.0.0.1:8081` and add two hostnames to
the existing Nginx reverse proxy. Cloudflare uses proxied A records for edge TLS,
while Nginx uses a Let's Encrypt certificate for encrypted edge-to-origin
traffic. The Atlas Nginx virtual host accepts origin requests only from the
published Cloudflare networks and localhost.

Nginx overwrites `CF-Connecting-IP` and forwarding headers before proxying. This
preserves the real client address for login throttling without trusting headers
from arbitrary origin callers. PostgreSQL remains on an internal Docker network.

## Consequences

- Atlas has two core containers: the Go application and PostGIS.
- No tunnel connector or tunnel credential is required.
- Atlas reuses ports 80/443 instead of competing for them.
- The application port and database are not reachable from remote hosts.
- Public availability still depends on Cloudflare DNS/proxy, but not on a tunnel.
- Operators must keep the Cloudflare source allowlist current and renew the
  origin certificate; Certbot installs a system timer for renewal.
- Self-hosters may use a different reverse proxy if they preserve loopback-only
  application exposure, TLS, forwarded-header trust, and database isolation.

## Rejected alternatives

- A second Caddy listener: ports 80/443 are already owned by the shared Nginx.
- DNS-only A records with a second proxy: duplicates the host ingress layer.
- Publishing `8081` publicly: bypasses TLS, origin filtering, and shared routing.
