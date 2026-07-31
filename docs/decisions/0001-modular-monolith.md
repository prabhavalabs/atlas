# ADR 0001: Use a Go modular monolith

- Status: Proposed
- Date: 2026-07-31

## Context

The product needs ingestion, correlation, editorial workflows, public reads, admin commands, scheduling, and auditing. A solo maintainer must run it on one VPS. Earlier experience showed that shared-database services plus webhooks and overlapping pipelines create partial success states that are difficult to reproduce.

## Decision

Build one Go 1.26 codebase and deployable application. Internal vertical modules own their tables and communicate through explicit interfaces/application services. A command mode can run API, worker, migrations, import, and administration tasks from the same image.

## Consequences

- Domain changes and job creation can be atomic.
- Local development, deployment, observability, and upgrades are simpler.
- Module boundaries are enforced by package and SQL ownership rather than network hops.
- Independent scaling is limited initially, which is acceptable for expected load.
- Extraction requires measured need and a new ADR.

## Rejected alternatives

- Microservices: excessive deployment, consistency, and tracing cost for one maintainer.
- Serverless functions: provider coupling, fragmented scheduling, and difficult local reproduction.
- Python primary backend: productive for experimentation but a larger runtime/dependency footprint than desired for this long-lived operational core.
