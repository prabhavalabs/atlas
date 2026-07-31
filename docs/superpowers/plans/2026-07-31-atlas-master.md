# Atlas master implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Use `superpowers:test-driven-development` for code tasks and `superpowers:verification-before-completion` before claiming a phase complete.

**Goal:** Deliver an account-free public disaster monitor and auditable administration portal that runs reliably on one VPS at minimal cost.

**Architecture:** A Go 1.26 modular monolith owns PostgreSQL/PostGIS, source ingestion, durable jobs, event correlation, editorial state, local admin identity, and REST APIs. One React/Vite/TypeScript workspace provides code-split public and admin route shells. Caddy, the app image, and PostGIS form the three-container core deployment.

**Tech stack:** Go 1.26, chi, Huma, pgx, sqlc, Goose, PostgreSQL/PostGIS, React, Vite, TypeScript, TanStack Router/Query, Zustand, shadcn/ui, Tailwind CSS, MapLibre GL JS, Vitest, Testing Library, MSW, Playwright, Docker Compose, Caddy.

## Decision gates before Task 1

- [x] Approve the product and repository identity: Atlas at `prabhavalabs/atlas`.
- [x] Select Apache-2.0 for original project code.
- [ ] Complete legal/trademark and domain review before public launch.
- [ ] Confirm target VPS CPU/RAM/storage architecture and backup provider.
- [ ] Approve four primary public/admin responsive designs and their degraded states.
- [ ] Record source terms/attribution for every enabled MVP adapter.

## Intended repository shape

```text
.
├── cmd/atlas/                 # server, worker, migrate, admin, import commands
├── internal/
│   ├── platform/              # config, database, HTTP, logging, metrics, clock
│   ├── jobs/
│   ├── source/
│   ├── observation/
│   ├── event/
│   ├── editorial/
│   ├── identity/
│   ├── audit/
│   ├── publicapi/
│   └── adminapi/
├── migrations/
├── queries/                   # sqlc SQL grouped by owning module
├── web/                       # React/Vite application
├── testdata/sources/          # licensed/redacted source fixtures
├── deploy/
│   ├── compose.yaml
│   ├── Caddyfile
│   └── scripts/
├── docs/
├── api/openapi.yaml           # generated and checked in
├── Dockerfile
├── Makefile
├── go.mod
└── sqlc.yaml
```

## Delivery sequence

1. Execute [foundation plan](2026-07-31-foundation.md).
2. Execute [ingestion plan](2026-07-31-ingestion.md).
3. Execute [event intelligence plan](2026-07-31-event-intelligence.md).
4. Execute [public web plan](2026-07-31-public-web.md).
5. Execute [admin portal plan](2026-07-31-admin-portal.md).
6. Execute [operations plan](2026-07-31-operations.md).

## Cross-plan invariants

- No network enrichment changes an unknown value to zero.
- One internal module owns every table and its SQL.
- Jobs are created transactionally with the domain state they advance.
- Public APIs expose only published revisions.
- Every public factual timeline/report claim has a citation.
- AI credentials are unnecessary for all acceptance tests.
- Public routes never require or redirect to admin identity.
- Source outage tests preserve last-known-good public data.
- Every migration is tested from empty and the previous release schema.
- A complete system uses no more than the three core containers.

## Phase verification commands

Run from repository root after every plan:

```bash
go test ./...
go vet ./...
go test -race ./internal/...
npm --prefix web run lint
npm --prefix web run typecheck
npm --prefix web run test:run
npm --prefix web run build
docker compose -f deploy/compose.yaml config --quiet
```

Before beta release also run:

```bash
make test-integration
make test-e2e
make test-restore
make scan
```

## Definition of done

- All subsystem plans and release gates pass.
- A clean VPS can be installed from documentation without private services.
- Fixture ingestion and live shadow ingestion produce explainable, reviewable events.
- An admin can publish and correct a cited report; an anonymous user can inspect every revision/source.
- Source, map, model, and backup outages have verified degraded behavior.
- Costs and enabled external dependencies are visible and bounded.
