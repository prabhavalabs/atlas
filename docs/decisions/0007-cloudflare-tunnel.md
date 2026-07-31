# ADR 0007: Use outbound-only Cloudflare Tunnel ingress

- Status: Accepted
- Date: 2026-07-31

## Context

The target VPS already hosts unrelated services and a host-level Nginx process.
Adding another public reverse proxy would create port ownership and certificate
coordination risk. Atlas needs two HTTPS hostnames, low operational cost, and no
public database or application port.

## Decision

Run a pinned `cloudflared` container beside the app and PostGIS. Cloudflare owns
public DNS/TLS and routes both Atlas hostnames through an outbound-only tunnel to
`http://app:8081`. Store the remotely managed tunnel token in a Compose secret
file and pass its path through `TUNNEL_TOKEN_FILE`.

The Go application serves both APIs and built Vite assets. The data network is
internal, the edge network contains only app and cloudflared, and Compose exposes
no Atlas host port.

## Consequences

- Atlas cannot collide with existing listeners on ports 80 or 443.
- TLS, basic DDoS protection, and CDN revalidation fit Cloudflare's free plan.
- Public availability depends on Cloudflare and outbound access to its edge.
- Self-hosters who avoid Cloudflare may place any TLS reverse proxy in front of
  `app:8081`; application and data boundaries do not depend on Cloudflare APIs.
- The token can start the tunnel and therefore requires file permissions,
  rotation, and repository-secret handling equivalent to a production credential.

## Rejected alternatives

- A second Caddy/Nginx listener on the host: conflicts with existing port ownership
  and expands shared-host coordination.
- Publishing an application port directly: exposes the VPS origin and adds manual
  TLS/firewall work.
- Kubernetes ingress: unjustified operational overhead for one application and
  one maintainer.
