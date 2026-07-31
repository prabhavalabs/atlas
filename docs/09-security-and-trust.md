# Security, safety, and public trust

## Trust model

The public may rely on the platform during stressful, fast-changing events. Security includes both conventional application security and protection against misleading content.

Trust boundaries are:

- untrusted upstream payloads and linked web content;
- anonymous public clients;
- authenticated administrators;
- optional model providers;
- database and backup storage;
- deployment and source credentials.

## Principal threats and controls

| Threat | Controls |
|---|---|
| Malformed or malicious feed payload | Strict schemas, size/time limits, safe XML parser settings, no entity expansion, fixture fuzzing, quarantine |
| Source spoofing or redirect to unsafe host | HTTPS, allowlisted hosts, redirect validation, DNS/IP checks where server-side URL fetching exists |
| SSRF from admin-submitted URL | Resolve and block private/link-local ranges, re-check redirects, restrict ports/protocols, egress allowlist where practical |
| Stored XSS in source text | Store plain text, sanitize any allowed markup on server, CSP, no arbitrary source HTML rendering |
| SQL injection | `sqlc`/parameterized queries; no user-composed SQL |
| Admin account takeover | Argon2id, secure HttpOnly SameSite cookies, CSRF tokens, rate limiting, optional TOTP/WebAuthn after MVP |
| Unauthorized editorial overwrite | Server-side authorization, optimistic revision checks, immutable audit trail |
| Accidental or malicious false publication | Evidence/citation validation, preview, publish permission, correction/unpublish flow |
| Prompt injection from evidence | Model has no tools or network, evidence delimiters, strict output schema, post-validation, human review |
| Dependency or image compromise | Locked dependencies, automated updates, SBOM, provenance/signing, vulnerability scan, minimal images |
| Credential exposure | Environment/secret files outside Git, startup redaction, least privilege, secret scan, rotation runbook |
| Data loss/ransomware | Encrypted off-site backups, immutable/versioned target where available, monthly restore drill |
| Traffic exhaustion | Cloudflare edge controls, bounded application inputs, database timeouts, cached public reads |

## Administration authentication

MVP administration is local to the installation; there is no third-party identity dependency. The bootstrap CLI creates the first administrator with a password read from standard input. Passwords use Argon2id. Sessions and CSRF secrets are independent random values and only their hashes are stored in the database. The session cookie is HttpOnly; a same-site readable CSRF cookie restores double-submit protection after a page reload. Sessions are expiring and revocable.

Initial roles are `administrator`, `editor`, and `viewer`:

- editor: review evidence, edit events/timelines/reports, request generation;
- administrator: editor capabilities plus publication and operational management;
- viewer: read-only administrative visibility.

Single-user deployments can assign `administrator`; the authorization boundary still exists and is tested.

## Public API protection

- Read-only public methods with bounded pagination and filter complexity.
- Per-IP/token-bucket rate limits that avoid retaining raw IPs longer than operationally necessary.
- Cache validators and short edge/browser caching for popular views.
- Maximum geometry vertex/payload sizes and server-side simplification.
- Query timeouts and statement timeouts.
- No CORS wildcard for admin APIs; public CORS is explicit and configurable.
- Security headers: CSP, HSTS, `nosniff`, restrictive permissions policy, and frame ancestors.

## Data minimization

Public monitoring requires little personal data. The system does not collect public accounts, ad identifiers, or precise visitor location by default. Optional “near me” filtering stays in the browser unless a user explicitly submits a coordinate query; it is not logged at full precision.

Administrator emails/usernames and audit identity are retained for operation. Source documents can contain personal information; retain and expose only what is necessary and lawful. A documented correction/removal channel exists before launch.

## Content safety rules

- Official alert text is attributed and timestamped.
- The platform never rewrites an advisory into its own imperative instruction.
- Casualties, missing persons, displaced population, and damage have source, time, status, and range.
- Conflicts are visible until resolved.
- Automated estimates are labeled and visually distinct from observed impact.
- Test, exercise, draft, and system messages are quarantined.
- A source outage cannot change an event to a reassuring state.
- Reports identify unknowns and coverage limitations.

## Security verification before public launch

- Threat-model review updated from implemented routes and deployment.
- Dependency, container, and secret scans pass.
- Admin authentication/CSRF/session tests pass.
- SSRF and stored-XSS tests pass against hostile fixtures.
- Backup restore and administrator recovery are demonstrated.
- Rate and resource exhaustion tests meet the VPS budget.
- Public vulnerability reporting policy and security contact are published.
