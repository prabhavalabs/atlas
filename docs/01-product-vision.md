# Product vision

## The promise

Atlas helps a person answer four questions quickly:

1. What significant disasters are active now?
2. What is known, when was it last verified, and by whom?
3. What changed since the previous update?
4. Where can I inspect the original evidence?

It is a monitoring and synthesis tool, not an emergency alerting authority. Every public screen must make that distinction clear and direct people to local authorities for life-safety decisions.

## Intended users

### Public monitor

Journalists, researchers, humanitarian workers, community organizations, and interested members of the public need a fast global overview without an account or setup. They value freshness, citations, filters, and a comprehensible history more than an exhaustive set of controls.

### Editorial administrator

A maintainer or trusted volunteer needs to resolve duplicate events, inspect source material, correct facts, edit a report, and publish with a durable audit trail. The system should reduce review work, not hide it behind an autonomous agent.

### Self-hosting maintainer

An open-source adopter needs a documented deployment, predictable resource use, source adapters that can fail independently, straightforward backups, and a system that still works without paid API keys.

## Product principles

- **Public by default.** Reading never requires an account.
- **Evidence before prose.** A claim is stored with its source before it appears in a summary.
- **Freshness is visible.** “Last checked,” “last changed,” and stale-source state are distinct.
- **Unknown is a valid value.** Missing impact data must not become a zero or imply safety.
- **Automation proposes; people publish.** High-impact editorial actions remain reviewable and reversible.
- **The useful path has no AI dependency.** Collection, correlation, timeline construction, and template reports are deterministic.
- **Degraded is better than blank.** Previously verified data remains available with a visible stale marker during upstream outages.
- **One maintainer can operate it.** A feature that adds a service, credential, or datastore must justify its ongoing operational cost.
- **Open-source is an operational requirement.** A new installation must be useful without proprietary infrastructure.

## Success measures

The initial six-month measures are deliberately operational:

| Measure | Target |
|---|---|
| Public availability | 99.5% monthly on a single VPS |
| Cached public API latency | p95 under 500 ms from the VPS region |
| Source outage behavior | Last-known-good data remains readable; source marked stale |
| Citation coverage | 100% of published factual timeline entries have at least one source |
| Report traceability | 100% of published report sections link to evidence records |
| Recovery point objective | 24 hours initially |
| Recovery time objective | 4 hours initially |
| Core deployment | No more than three long-running containers |
| Unplanned monthly spend | Zero; all optional services have hard limits or require explicit enablement |

## Product boundaries

The platform presents an aggregated view and may make mistakes. It must not:

- issue evacuation or safety instructions in its own voice;
- represent model output as an official statement;
- infer casualties or damage from weak signals;
- hide conflicting source values;
- silently merge ambiguous events;
- republish full copyrighted articles;
- depend on a social-media engagement score as evidence of factual severity.

## Identity

The product name is **Atlas**, with the endorsement **“by Prabhava Labs”** where organizational context is useful. The public tagline is **“Global disaster intelligence, built in the open.”** Atlas is not an acronym and should not be styled as `ATLAS`. The repository, Go module, command, and container identity use `prabhavalabs/atlas`, `github.com/prabhavalabs/atlas`, `atlas`, and `ghcr.io/prabhavalabs/atlas` respectively. Complete legal/trademark and domain clearance before the first public release.
