# Risks and mitigations

| Risk | Likelihood | Impact | Mitigation and trigger |
|---|---|---|---|
| Upstream source outage or schema change | High | High | Fixtures, contract monitors, conditional fetch, last-known-good snapshot, per-source circuit breaker; alert on freshness SLO |
| False merge of separate events | Medium | High | Type-specific deterministic scoring, ambiguity queue, reversible merge/split, stored features; review threshold calibrated on golden set |
| Duplicate events remain separate | Medium | Medium | Candidate review queue and alias/source-crosslink logic; measure unresolved duplicate rate |
| Preliminary data presented as fact | Medium | High | Status/confidence/source distinctions, public labels, publication checks, correction history |
| Generated report invents or alters facts | Medium if enabled | High | Bounded evidence, no tools, strict citations/schema, numeric validation, human publish gate, template fallback |
| Copyright violation from news ingestion | Medium | High | Store metadata/short excerpt/link by default, source retention policy, removal process, no wholesale article crawling |
| Map service becomes unavailable or changes terms | Medium | Medium | Configurable provider and local low-zoom PMTiles fallback; attribution tests |
| PostGIS/database loss on one VPS | Low–medium | Critical | Encrypted off-site backups, monthly restore drill, disk alerts, documented rebuild |
| Maintainer overload | High | High | MVP scope lock, source budget, automated health, ADR complexity budget, contribution templates |
| Open-source component has unclear license | Medium | High | Dependency/data license inventory, SBOM/NOTICE, no reuse without explicit license |
| Population estimate misread as confirmed impact | Medium | High | Label as modeled exposure, separate schema/UI, methodology and confidence, never convert to casualties |
| Admin account compromise | Low–medium | High | Secure sessions, strong password hashing, rate limit, audit, revocation, TOTP/WebAuthn post-MVP |
| Feed content exploits parser/browser | Medium | High | Safe parsers, size limits, sanitization, CSP, no source HTML execution, fuzz tests |
| Global coverage creates false completeness | High | High | Coverage/source-health map, regional gaps, freshness indicators, “not monitored” distinct from “no event” |
| Disaster traffic spike exhausts VPS | Medium | High | Cacheable public reads, Caddy limits, bounded geometry, load test, emergency static snapshot mode |
| Configuration/migration drift | Medium | High | Generated `sqlc`, migration checks from empty and previous release, schema-contract integration tests, atomic releases |
| Full-text search insufficient | Low initially | Low | Measure failed/abandoned searches; improve aliases/trigrams before adding semantic infrastructure |
| Atlas name conflicts or is confused with another product | Unknown | Medium | Trademark/domain review before public launch; always pair with disaster-monitoring description and Prabhava Labs endorsement |

## Risk acceptance rules

- No critical/high risk can be silently accepted; its owner and rationale are recorded in an issue or ADR.
- Source coverage limitations appear in the product, not only in internal monitoring.
- Model, source, and map provider terms are rechecked before each major release.
- The MVP launch is a monitored beta and explicitly not an emergency notification service.

## Future improvements, ordered by evidence

1. More official CAP adapters and multilingual normalization.
2. Optional population exposure with a properly licensed, benchmarked component.
3. PWA caching and low-bandwidth text view.
4. Contributor/editor workflow and stronger identity such as WebAuthn.
5. Notifications/feeds for published updates, with careful delivery semantics.
6. Additional authoritative hazard layers such as flood/fire products.
7. Read replicas or additional workers after observed load.
8. Semantic retrieval only if measured search failures justify it.
