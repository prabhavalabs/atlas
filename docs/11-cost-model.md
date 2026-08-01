# Cost model

Prices were checked on 2026-07-31 and are planning estimates, not commitments. Provider pricing and taxes vary by region and can change.

## Baseline with the existing VPS

| Item | Monthly estimate | Notes |
|---|---:|---|
| Existing VPS | $0 incremental | Requires roughly 4 vCPU, 8 GB RAM, 80–160 GB SSD |
| Data-source APIs | $0 | MVP sources are free/public; ReliefWeb app name approval and NASA FIRMS key are administrative, not paid |
| Database, proxy, application | $0 | Open-source application/PostGIS/Nginx; Cloudflare proxy and DNS use the free plan |
| Map tiles | $0 initially | OpenFreeMap public service plus self-hosted low-zoom fallback; no SLA |
| Off-site backup | $0–$1 | Small B2 usage can fit its current first-10-GB allowance; storage starts around $6.95/TB/month beyond that |
| AI | $0 | Deterministic reports; optional generation disabled |
| Uptime monitor | $0 | Free tier or self-hosted check |
| Domain | about $1–$2 amortized | Typically $12–$24/year, registrar/TLD dependent |
| **Expected incremental total** | **$1–$4/month** | Mostly domain and backup growth |

## New-host reference

An ARM shared-vCPU VPS around 4 vCPU/8 GB is the cost-efficient target. As a current point of reference, Hetzner's June 2026 pricing lists CAX31 at €20.99/month before VAT and IPv4. Provider choice should also consider backup region, support, bandwidth, and architecture compatibility—not price alone.

Expected core budget on a new host: **$22–$40/month**, including tax/IP variation, modest backups, and domain.

## Map delivery scenarios

| Scenario | Cost | Trade-off |
|---|---:|---|
| OpenFreeMap + local low-zoom fallback | $0 | No service-level agreement; sufficient for early public traffic |
| Protomaps hosted API | Free for non-commercial; sponsorship for commercial | Soft limit documented around 1M tile requests/month; verify current sponsor tier |
| Self-host global PMTiles | Storage/egress dependent | Full planet roughly 120 GB; object storage/CDN preferred over VPS disk |
| MapTiler Flex | $30/month currently | Includes 25k sessions and 500k requests; commercial hosted option |

The public map provider is configurable so cost or terms do not force an application rewrite.

## Report-generation scenarios

| Scenario | Direct cost | Infrastructure/privacy |
|---|---:|---|
| Deterministic templates | $0 | Core default; strongest reproducibility |
| OpenRouter free model | $0 within current quota | Evaluation/manual fallback only; model availability and limits can change |
| OpenRouter paid model | Model-dependent, usage based | Pin an approved model/provider and enforce a smaller local hard budget |
| Remote free tier | $0 within quota | Quota not an SLA; free-tier submissions may be used for provider improvement |
| Remote paid, operator-triggered | Usually cents to low single digits/month at MVP volume | Enforce local token/request cap; verify current price and data terms |
| Local Ollama | $0 per token | Requires 16–32 GB profile for useful models; power/CPU opportunity cost |

For scale estimation, if 300 report revisions/month each use 12k input and 1.5k output tokens, that is 3.6M input and 0.45M output tokens. At a hypothetical $1.50/M input and $9/M output, usage is about $9.45/month. This is a scenario calculation, not a provider promise. Section-only generation and caching reduce it.

## Cost controls

- No external service is enabled merely by adding a credential; an administrator also sets a monthly budget.
- Hard local request/token caps stop optional AI before spend exceeds the configured ceiling.
- HTTP caching and conditional requests reduce source and tile traffic.
- Raw bodies have retention/size caps.
- Storage growth, backup size, and database table size are visible in admin health.
- General web search, browser proxies, commercial geocoding, and paid queues are absent from MVP.
- A quarterly cost report compares actual spend and resource utilization with this model.

## Cost threshold decisions

- Under 5,000 monthly map sessions: stay on free/open map path.
- Sustained public usage or map-provider instability: sponsor/contract a hosted provider or self-host object-storage PMTiles.
- Database exceeds 60% disk or backup window exceeds target: tune retention/indexes before resizing.
- Optional services push recurring cost above $50/month: publish a cost ADR and user-value evidence.
