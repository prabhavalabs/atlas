# Design lessons from the predecessor review

This public document records reusable lessons from an authorized review of a private predecessor system. It intentionally excludes names, private repository paths, customer information, internal URLs, credentials, screenshots, source code, and infrastructure identifiers.

## What worked

- A three-layer event model—raw source item, normalized observation, canonical event—made replay and source attribution possible.
- Source adapters with a shared normalization contract made additional official feeds easier to introduce.
- Deterministic spatial and temporal correlation was more explainable than model-led merging.
- PostGIS was a good fit for event geometry, proximity, tracks, and affected areas.
- Preserving report versions and showing citations made editorial correction possible.
- Explicit preliminary/predicted states helped avoid presenting automatic data as confirmed fact.
- Moving ingestion from a serverless chain to a small Go service reduced latency and operational surface.
- Source-specific fallback scoring showed that most critical event classification can be computed from already collected data.

## What became unnecessarily complex

- Multiple runtimes and services shared a database while also coordinating through webhooks and background workflows.
- The same event was scored, enriched, scheduled, summarized, and notified by several overlapping pipelines.
- General web search, browser crawling, document extraction, embeddings, agent tools, and several model providers became part of the normal reporting path.
- Notifications depended on precise ordering across ingestion, priority changes, lifecycle changes, schedulers, and external workers.
- Multiple job systems, caches, queues, and deployment generations coexisted during migration.
- Public monitoring, organization administration, incident response, team management, and AI experimentation competed in one information architecture.
- Documentation described several architectural generations at once, making the deployed truth difficult to identify.
- Frontend and backend surface area grew much faster than automated coverage of core event workflows.

## Failure patterns to prevent

### Enrichment must not gate core classification

An upstream detail request can time out even when enough data already exists to classify an event. A network failure must produce `unknown` enrichment and a retry, not a severity of zero or a dropped event.

### Cross-process write ownership causes hidden coupling

If one service writes another service's tables and then invokes a webhook, database success and workflow success can diverge. In the new platform, one module owns each table and asynchronous work is recorded transactionally in the same database.

### State transitions need invariants

Rapid source updates can flap severity or lifecycle values. State changes need hysteresis, minimum evidence, source precedence, and tests for escalation, downgrade, cancellation, correction, and retraction.

### Schema and code must be one release unit

Integration tests in the reviewed Go ingestion path exposed enum drift between application values and the database. Migrations, generated query code, fixtures, and contract tests must ship together.

### A report is only as reliable as its evidence loader

Collecting a document is insufficient if later filters silently omit it. Each report section must expose the exact evidence set it received, and uploaded/manual sources must have explicit inclusion rules.

### More fallbacks can create more failure modes

Search provider, browser mode, proxy, and model fallbacks expand the test matrix. Fallbacks belong at clear boundaries and must degrade to a useful deterministic result.

## Rules adopted for Atlas

1. One Go process owns the domain and database writes.
2. One Postgres database is the durable coordination mechanism.
3. Source ingestion is feed/API first; browser crawling is outside the core product.
4. AI is an optional renderer, not a collector, correlator, scorer, or publisher.
5. Every source exposes freshness, last success, error streak, and last-known-good snapshot.
6. Core severity and priority use local, deterministic inputs with `unknown` as a first-class state.
7. Public and administration experiences have different route shells and permissions.
8. The MVP has an explicit complexity budget: three containers, one datastore, no external queue, no cache service, and no vector database.
9. Architecture decisions have one current document and superseded decisions are marked explicitly.
10. A feature is not complete until its failure, stale, empty, and recovery states are tested.
