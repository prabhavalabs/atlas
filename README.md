# Atlas

**Global disaster intelligence, built in the open.**

Atlas is an open-source platform from Prabhava Labs for monitoring disasters around the world. It is designed to aggregate authoritative alerts, humanitarian updates, maps, news, event timelines, and cited situation reports into a public experience that works without an account.

> [!IMPORTANT]
> Atlas is an early foundation release. Its source coverage is not yet sufficient for life-safety use, and it must not be treated as an emergency alerting authority.

## Product experience

Atlas will provide two deliberately separate interfaces:

- **Public platform:** account-free active-event map and list, filters, search, event narratives, timelines, source health, news, and situation reports.
- **Administration portal:** evidence review, event correlation and correction, timeline/report editing, source health, publication, and an immutable audit trail.

Every published factual claim is expected to retain a source. Preliminary, automatic, estimated, conflicting, and unknown information remain visibly distinct.

## Architecture direction

- Go 1.26 modular monolith
- PostgreSQL with PostGIS
- PostgreSQL-backed durable jobs; no external queue or Redis
- React, Vite, TypeScript, TanStack Query, Zustand, and shadcn/ui
- MapLibre GL JS with configurable map delivery
- Deterministic timelines and reports that work without AI
- Optional evidence-bounded, human-reviewed AI report rendering
- Three-container VPS deployment: application, PostGIS, and Cloudflare Tunnel

The project intentionally excludes microservice sprawl, browser crawling, general web search, a vector database, and agent-led publication from its initial complexity budget.

## Documentation

Start with the [documentation index](docs/README.md).

- [Product vision](docs/01-product-vision.md)
- [Experience design](docs/02-experience-design.md)
- [Architecture](docs/05-architecture.md)
- [Data model](docs/06-data-model.md)
- [Data sources](docs/07-data-sources.md)
- [Reports and optional AI](docs/08-reports-and-ai.md)
- [Deployment and operations](docs/10-deployment-and-operations.md)
- [Costs](docs/11-cost-model.md)
- [Roadmap](docs/13-roadmap.md)
- [Master implementation plan](docs/superpowers/plans/2026-07-31-atlas-master.md)

## Foundation status

The repository contains a deployable walking skeleton: embedded migrations, a fixture-backed event path, cacheable public APIs, local administrator sessions and RBAC, audited optimistic updates, a public list/map experience, an administration sign-in/review view, durable PostgreSQL jobs, a generated OpenAPI client, and automated container delivery.

Real-source ingestion, event correlation, full timelines and reports, editorial publication workflows, backups, and production hardening remain roadmap work. Current behavior and limitations are tracked in [the roadmap](docs/13-roadmap.md).

## Local development

Requirements are Go 1.26.5, Node.js 24, pnpm 11.11, Docker, and Docker Compose.

```sh
make bootstrap
make test
make build
```

Run PostGIS locally, export `ATLAS_DATABASE_URL`, `ATLAS_PUBLIC_URL`, and
`ATLAS_API_URL`, then start the application with `go run ./cmd/atlas serve`.
The production deployment template is `.env.example`; the Go process does not
implicitly load dotenv files. Operator commands are `migrate`,
`import-fixture`, `create-admin`, `healthcheck`, and `version`.

## Contributing

Atlas welcomes careful contributions from the disaster-management, humanitarian, geospatial, journalism, accessibility, and open-source communities. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening an issue or pull request.

Incorrect disaster information should be reported with the dedicated correction issue template. Security vulnerabilities must be reported privately as described in [SECURITY.md](SECURITY.md).

## Safety

Atlas aggregates third-party information and may be incomplete, delayed, or incorrect. It does not replace local authorities, emergency services, official warnings, or professional judgment. Life-safety decisions must be based on information from the responsible local or national authority.

## License

Atlas is licensed under the [Apache License 2.0](LICENSE). Data and third-party components retain their own licenses and attribution requirements; these are tracked in [NOTICE](NOTICE) and the source registry.
