# Reference project review

Reviewed on 2026-07-31. These projects are inspirations, not dependencies. Architectural ideas may be adapted; code is copied only when its license is compatible and required notices are preserved.

## World Monitor

- Repository: [koala73/worldmonitor](https://github.com/koala73/worldmonitor)
- License: AGPL-3.0-only

### Useful ideas

- Scheduled source collectors publish bounded snapshots and public readers receive cached last-known-good data.
- Responses expose fetch time and data-availability state instead of treating a source outage as an empty world.
- Disaster aggregation combines official sources such as GDACS, NASA EONET, USGS, NHC, and NASA FIRMS with source-specific filters.
- News ranking separates source tier, recency, severity signals, and corroboration.
- Feed ingestion validates dates, removes future or undated items, de-duplicates normalized titles, and uses a defined freshness floor.
- Expensive or probabilistic synthesis is bounded by deterministic evidence selection.

### What not to adopt

Its broad mission has grown into many providers, feeds, deployment targets, caches, clients, and intelligence domains. That is appropriate for its product, but too large for a single-maintainer disaster MVP. Atlas will reproduce last-known-good behavior in Postgres and use a much smaller source catalog.

AGPL code is not copied into the Apache-2.0 project. If that changes, the project license implications require a separate legal and architecture decision.

## GeoLibre

- Repository: [opengeos/GeoLibre](https://github.com/opengeos/GeoLibre)
- License: MIT

### Useful ideas

- MapLibre is the default rendering engine, while domain state remains plain and engine-independent.
- Heavy geospatial features are loaded on demand.
- PMTiles and Protomaps enable portable, self-hostable vector data.
- GeoJSON remains the interchange format at product boundaries.
- Accessibility, browser automation, and focused geospatial unit tests are treated as core quality work.

### What not to adopt

Atlas is not a browser GIS workbench. It does not need DuckDB-WASM, Cesium, desktop/mobile packaging, plugin execution, a notebook bridge, or general-purpose spatial processing in the first release. Spatial analysis remains on the server in PostGIS; the browser renders curated layers.

## Last 30 Days

- Repository: [mvanhorn/last30days-skill](https://github.com/mvanhorn/last30days-skill)
- License: MIT

### Useful ideas

- Collection, enrichment, ranking, and synthesis are distinct pipeline stages.
- Expensive work can be checkpointed and resumed with identity, provenance, and time-to-live validation.
- A fixed confidence floor is better than selecting the “best” weak result from a poor pool.
- “Nothing solid” is a valid, honest outcome.
- Enrichment failures downgrade a candidate instead of failing the entire run.
- Evidence clusters retain provenance and can support cumulative summaries across a defined time window.

### Adaptation for disaster timelines

Atlas replaces the fixed 30-day window with an event window from the first observation through the present. The evidence unit is an authoritative observation or attributed publication, not social engagement. Timeline candidates are generated from material changes, de-duplicated, assigned confidence, and reviewed. The report describes the full event history while emphasizing changes since the previous published revision.

### What not to adopt

The runtime will not depend on browser cookies, social-network scraping, engagement scoring, multi-agent judgment, or a general web-research harness. These inputs are volatile and difficult for open-source operators to reproduce.

## Existing Prabhava Labs geospatial components

### Population and reverse-geocoding service

An existing Prabhava Labs service demonstrates fast PostGIS-backed reverse geocoding and population exposure analysis using WorldPop, GeoNames, and Natural Earth. It is technically valuable, but the reviewed public repository does not currently contain a detectable license file. Public visibility is not permission to reuse: GitHub's licensing guidance notes that default copyright applies without a license.

Before reuse:

1. confirm copyright ownership and all contributor permissions;
2. add an explicit compatible software license;
3. document dataset versions, attribution, and redistribution terms;
4. benchmark storage and memory on the target VPS;
5. separate approximate exposure estimates from verified impact figures.

The full population-cell deployment is optional after MVP because its database footprint is disproportionate for a small VPS. MVP reverse geocoding uses a compact subset of GeoNames cities plus Natural Earth administrative boundaries in the primary database. The larger service can later run as an optional profile with a stable HTTP interface.

### React/Vite map application

An existing MIT-licensed Prabhava Labs React/Vite application provides reusable experience with React 19, Vite, TypeScript, shadcn/ui, MapLibre, Zustand, Tailwind, and PMTiles. Generic map-shell patterns may be adapted with attribution. Product-specific content and visual identity are not carried into Atlas. TanStack Query and the accessibility/test stack still need to be added.

## License compatibility rule

Apache-2.0, MIT, BSD, and compatible permissive code may be reused with notices. AGPL/GPL code is studied but not copied unless the entire distribution decision is reconsidered. Data licenses are tracked separately from software licenses in the source registry and NOTICE files.
