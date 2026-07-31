# Roadmap and scope

## Phase 0 — decisions and design (1–2 weeks)

Exit criteria:

- Atlas passes legal/domain review and its public endpoint is selected;
- Apache-2.0 notices and third-party/data attribution policy are confirmed;
- public overview, event page, admin queue, and event workspace have reviewed responsive designs for failure/stale/empty states;
- source terms for the MVP are recorded;
- target VPS resources and backup destination are confirmed;
- architecture and plans in this directory are accepted.

No feature code should precede these gates.

## Phase 1 — walking skeleton (2 weeks)

Deliver a deployable vertical slice:

- Go application, migrations, PostGIS, jobs, public/admin route shells;
- React/Vite workspace with design tokens and generated API client;
- one fixture-backed source adapter;
- raw observation to canonical event to public list/detail;
- local admin login and audit event;
- Compose deployment, CI, backup/restore smoke test.

The aim is deployment confidence, not source breadth.

## Phase 2 — reliable core ingestion (3–4 weeks)

- GDACS, USGS, NASA EONET, and ReliefWeb adapters;
- source registry, health/freshness UI, last-known-good behavior;
- deterministic correlation, aliases, revisions, merge/split review;
- event lifecycle and priority rules;
- compact GeoNames/Natural Earth import and reverse geocoding;
- robust fixture, replay, fuzz, and schema-contract tests.

Exit: an upstream outage cannot erase events or turn unknown into low severity.

## Phase 3 — public product (3 weeks)

- synchronized list/map global overview;
- shareable filters and search;
- accessible public event narrative, geometry, timeline, sources/news;
- responsive/mobile and low-bandwidth behavior;
- map provider configuration and low-zoom fallback;
- public methodology, coverage, correction, attribution, and API docs.

Exit: public use requires no account and core content is keyboard/screen-reader accessible.

## Phase 4 — editorial administration (3 weeks)

- review queue and event workspace;
- evidence linking, suppression, merge/split, revision comparison;
- timeline candidate review;
- structured deterministic report editor;
- publish/unpublish/correction workflow;
- source health and failed-job controls;
- admin audit and role authorization.

Exit: every public fact/report paragraph is traceable and corrections are reversible.

## Phase 5 — beta hardening (2–3 weeks)

- load, soak, chaos/failure, security, accessibility, and restore testing;
- operational dashboards/runbooks and retention enforcement;
- SBOM, signed images, vulnerability policy, contribution docs;
- real-source shadow run and correlation threshold calibration;
- deployment upgrade/rollback rehearsal;
- public beta release.

## Phase 6 — optional intelligence (after stable beta)

- operator-triggered section renderer with template/Ollama/remote providers;
- report evaluation golden set and budget enforcement;
- additional CAP/NHC/FIRMS adapters;
- optional population exposure profile;
- update feeds/notifications only after publication workflow proves stable.

## MVP definition

MVP includes four global/official data integrations, one CAP pilot, event list/map/detail, source health, deterministic timelines/reports, local admin review/publish, PostGIS search/correlation, Compose deployment, backups, and documented public methodology.

## Explicitly deferred

- multi-tenant organizations and complex roles;
- incident response/deployment/contact management;
- generic web search and browser crawling;
- PDF/document OCR pipelines;
- chat, RAG, embeddings, vector database, and agent tools;
- mobile/desktop native packages;
- dashboard widget customization;
- automatic social-media posting or notification matrices;
- multiple independent backend services.

## Resourcing estimate

For one experienced full-time maintainer, the beta plan is approximately 14–18 focused weeks plus source/legal review time. Part-time work should be planned as milestone-based rather than converted mechanically into calendar dates. The critical path is source semantics and editorial correctness, not UI scaffolding.

## Release gates

Each phase requires passing tests, docs updates, one-VPS deployment verification, no unexplained schema/API drift, and a review of complexity/cost changes. A phase is complete only when its failure and recovery behavior is demonstrated.
