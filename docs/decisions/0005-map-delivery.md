# ADR 0005: Use MapLibre with configurable tiles and a local overview fallback

- Status: Proposed
- Date: 2026-07-31

## Context

Maps are central, but commercial APIs can dominate cost and community OpenStreetMap tile servers are not a production SLA. A full global self-hosted high-zoom basemap is large for one VPS.

## Decision

Use MapLibre GL JS. Configure the style/tile provider at deployment. Start with an open hosted provider compliant with its terms, while serving a self-hosted low-zoom global PMTiles fallback sufficient for the overview. Move higher-detail PMTiles to object storage/CDN or a paid provider when usage requires it.

## Consequences

- No renderer lock-in and no Google Maps dependency.
- The global overview remains usable through provider failure.
- Attribution and provider health are product requirements.
- Street-level detail can degrade while event geometry/list remains available.

## Rejected alternatives

- Direct `tile.openstreetmap.org` production dependency: best-effort and capacity constrained.
- Full planet high-zoom archive on core VPS: roughly 120 GB plus update/egress burden.
- Proprietary renderer/provider coupling: recurring cost and migration risk.
