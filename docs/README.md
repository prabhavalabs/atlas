# Atlas planning documentation

- **Status:** architecture and implementation planning
- **Product name:** Atlas
- **Owner:** Prabhava Labs
- **Last reviewed:** 2026-07-31

Atlas is an open-source, public-first disaster monitoring platform from Prabhava Labs. It will combine authoritative alerts, humanitarian updates, maps, news, and cited situation reports without requiring a public account. The repository identity is `prabhavalabs/atlas`; legal and domain clearance remains a release gate.

This directory is the source of truth for the product and technical plan. No predecessor names, private repository identifiers, internal endpoints, customer information, or proprietary implementation details belong in this public documentation.

## Decision summary

| Area | Decision |
|---|---|
| Product shape | Public map and event narratives, plus a separate administration portal |
| Backend | Go 1.26 modular monolith with explicit internal modules |
| Data | PostgreSQL with PostGIS; append-only observations and versioned publications |
| Background work | PostgreSQL-backed jobs and advisory-lock scheduler; no Redis or external queue |
| Frontend | React, Vite, TypeScript, TanStack Query, lightweight Zustand UI state, shadcn/ui |
| Maps | MapLibre GL JS; configurable tile provider; low-zoom self-hosted PMTiles fallback |
| Search | PostgreSQL full-text and trigram search; no Elasticsearch or vector database |
| AI | Optional report renderer behind a provider interface; deterministic reports always work |
| Deployment | One VPS with Docker Compose, Caddy, one application image, and PostGIS |
| Core containers | Three: Caddy, application, database |
| License | Apache-2.0 |
| MVP sources | GDACS, USGS earthquakes, NASA EONET, ReliefWeb, and selected official RSS/Atom/CAP feeds |

## Experience and architecture

- [Product vision](01-product-vision.md)
- [Experience design](02-experience-design.md)
- [Design lessons](03-design-lessons.md)
- [Reference project review](04-reference-projects.md)
- [System architecture](05-architecture.md)
- [Data model and event lifecycle](06-data-model.md)
- [Data sources and licensing](07-data-sources.md)
- [Reports, timelines, and optional AI](08-reports-and-ai.md)
- [Security, safety, and public trust](09-security-and-trust.md)
- [Deployment and operations](10-deployment-and-operations.md)
- [Cost model](11-cost-model.md)
- [Risks and mitigations](12-risks.md)
- [Roadmap and scope](13-roadmap.md)
- [Testing and quality](14-testing-and-quality.md)

## Architecture decisions

- [ADR 0001: modular monolith](decisions/0001-modular-monolith.md)
- [ADR 0002: PostgreSQL-backed jobs](decisions/0002-postgres-jobs.md)
- [ADR 0003: evidence-first reporting](decisions/0003-evidence-first-reporting.md)
- [ADR 0004: public and admin route separation](decisions/0004-public-admin-separation.md)
- [ADR 0005: map delivery](decisions/0005-map-delivery.md)
- [ADR 0006: open-source license](decisions/0006-open-source-license.md)

## Implementation plans

Development should not begin until the foundation plan's decision gates are resolved. Plans are ordered, test-first, and independently reviewable:

1. [Master delivery plan](superpowers/plans/2026-07-31-atlas-master.md)
2. [Repository and application foundation](superpowers/plans/2026-07-31-foundation.md)
3. [Ingestion and source reliability](superpowers/plans/2026-07-31-ingestion.md)
4. [Event correlation, timelines, and reports](superpowers/plans/2026-07-31-event-intelligence.md)
5. [Public web application](superpowers/plans/2026-07-31-public-web.md)
6. [Administration portal](superpowers/plans/2026-07-31-admin-portal.md)
7. [VPS operations and release](superpowers/plans/2026-07-31-operations.md)

## Non-goals for the first release

The MVP does not include multi-tenancy, team management, a general web-search engine, browser-based crawling, chat/RAG, a vector database, mobile apps, desktop apps, custom dashboards, incident deployments, contact management, outbound alerting, or a microservice control plane. Those capabilities require evidence of user need before they can consume the project's complexity budget.
