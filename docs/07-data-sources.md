# Data sources and licensing

## Source policy

The source catalog is an allowlist. Every adapter requires a documented owner, authority scope, update cadence, authentication method, terms/license URL, attribution, retention rule, rate limit, failure mode, fixture set, and maintainer contact. A source cannot be enabled in production until this record is reviewed.

Official alerts and humanitarian reports are preferred. News enriches an event but never creates a high-severity canonical fact by itself. Social sources are outside the MVP.

## MVP sources

| Source | Purpose | Access/cost | Important constraints | Phase |
|---|---|---|---|---|
| [GDACS](https://www.gdacs.org/Documents/2025/GDACS_API_quickstart_v1.pdf) | Global earthquakes, floods, cyclones, volcanoes, drought and wildfire alerts/impact models | Free API, GeoJSON, RSS | Attribute GDACS; automated estimates require validation and do not replace local authorities | MVP |
| [USGS Earthquake Hazards](https://earthquake.usgs.gov/earthquakes/feed/) | Authoritative real-time earthquake feed and event details | Free GeoJSON feeds | Prefer real-time feeds over catalog queries; preserve updates/deletions and magnitude type | MVP |
| [NASA EONET v3](https://eonet.gsfc.nasa.gov/docs/v3) | Curated global natural-event metadata and geometry | Open API, no key documented | Curated reference, not an alerting authority; preserve its underlying source links | MVP |
| [ReliefWeb API v2](https://apidoc.reliefweb.int/) | Humanitarian reports, disaster groupings, response updates | Free read API; approved `appname` required since 2025-11-01 | Partner content may be copyrighted; store metadata/excerpt/link unless rights permit more | MVP |
| Official RSS/Atom/CAP registry | Local/national authority alerts and situation updates | Usually free | One reviewed configuration per authority; CAP test/exercise messages must never be public incidents | MVP pilot |

GDACS terms explicitly state that automatic alerts can contain errors and are not a substitute for local/national authorities. That warning is reflected in the UI and data model.

## Early post-MVP sources

| Source | Purpose | Access/cost | Decision |
|---|---|---|---|
| [NWS alerts API](https://www.weather.gov/documentation/services-web-api) | US weather and hydrological CAP alerts | Free, no API fee; identifying User-Agent required; rate limits | Add as the first CAP reference adapter; poll no more frequently than guidance |
| [National Hurricane Center GIS/RSS](https://www.nhc.noaa.gov/gis/rss.php) | Atlantic/eastern/central Pacific cyclone tracks, cones, watches | Free feeds | Valuable authoritative detail; service is experimental/no delivery guarantee, so retain last good geometry |
| [NASA FIRMS](https://firms.modaps.eosdis.nasa.gov/api/map_key/) | Near-real-time active fire detections | Free MAP_KEY; 5,000 transactions per 10 minutes | Detection is a signal, not a wildfire event; aggregate carefully and corroborate |
| [Copernicus Emergency Management Service](https://emergency.copernicus.eu/) | European/global mapping and flood/fire products | Mostly open with product-specific access | Add only after exact machine interface and redistribution terms are reviewed |
| Additional national CAP authorities | Regional official warnings | Source-specific | Add from a maintained registry with fixtures and language metadata |

## News and publication strategy

Start with:

1. ReliefWeb reports linked by disaster and geography;
2. RSS/Atom from official emergency agencies and reputable humanitarian organizations;
3. a small, manually reviewed publisher feed registry;
4. administrator-submitted URLs.

For each item, validate publication time, canonicalize URL, hash normalized title, keep publisher/source tier, and cluster similar headlines. Rank with deterministic features: authoritative source class, event relevance, corroboration, freshness, and material change. Do not scrape search-result pages or make a paid general-search API part of routine ingestion.

A simple HTTP article fetcher may extract title, publication time, metadata, and a short excerpt when robots/terms allow. It has strict time/size/content-type limits and no JavaScript browser. Failed extraction leaves the original link usable.

## Geospatial reference data

| Dataset | Use | License/terms | MVP treatment |
|---|---|---|---|
| [Natural Earth](https://www.naturalearthdata.com/about/terms-of-use/) | Country/admin boundaries and small-scale labels | Public domain | Import selected generalized boundaries into PostGIS |
| [GeoNames](https://www.geonames.org/) | Place names and reverse-geocoding candidates | CC BY 4.0 | Import populated places/admin names; show attribution |
| [WorldPop](https://hub.worldpop.org/geodata/) | Approximate population exposure | Dataset-specific, commonly CC BY 4.0; some derived products ODbL | Optional profile only after verifying exact release license |
| OpenStreetMap-derived basemap | Geographic context | ODbL and provider-specific policy | Visible attribution; provider URL is configurable |

Dataset provenance includes exact release identifier, checksum, import date, license, and transformation. Border disputes and approximate centroids need explicit cartographic policy.

## Map tile options

Production must not rely on community `tile.openstreetmap.org`: its [tile policy](https://operations.osmfoundation.org/policies/tiles/) states that data is free but the tile service is capacity-limited, best-effort, and unsuitable for bulk/offline use.

Recommended sequence:

1. **Development/early public:** [OpenFreeMap](https://openfreemap.org/) public instance, which currently states no keys or request limits, with the understanding that it has no project-controlled SLA.
2. **Resilience fallback:** self-host a low-detail global PMTiles archive (for example, zoom 0–6 is about 60 MB in the [Protomaps guide](https://docs.protomaps.com/guide/getting-started)) so the global overview never becomes blank.
3. **Scale:** sponsor/use the Protomaps hosted API or place a current PMTiles archive on affordable object storage/CDN. A full planet archive is roughly 120 GB, so it is not copied onto the core VPS by default.
4. **Commercial hosted alternative:** MapTiler's free plan is limited to non-commercial use; its current paid entry plan is documented in the cost model.

The client reads a style URL from deployment configuration. Attribution remains visible and is tested.

## Adapter reliability requirements

- Conditional HTTP (`ETag`, `If-Modified-Since`) where supported.
- Identifying User-Agent and contact URL/email.
- Per-source timeouts, maximum bodies, request budgets, and concurrency.
- Versioned raw fixtures for normal, empty, malformed, revised, retracted, test, and rate-limited responses.
- Last successful fetch and last material change stored separately.
- Exponential backoff and circuit breaker with manual retry.
- A last-known-good snapshot retained after failure.
- Metrics for freshness, duration, item counts, error class, and change volume.
- Configured alert when a source exceeds its freshness objective.

## Source onboarding checklist

- [ ] Authority and geographic/hazard scope recorded.
- [ ] Terms, license, attribution, and content retention approved.
- [ ] Authentication secret is optional or documented with a free/low-cost path.
- [ ] Poll cadence and published rate guidance recorded.
- [ ] Stable identity and update/retraction semantics understood.
- [ ] Time, geometry, units, languages, and test/exercise messages handled.
- [ ] Fixtures and contract tests cover degradation.
- [ ] Public methodology page updated.
- [ ] Disable/rollback procedure tested.
