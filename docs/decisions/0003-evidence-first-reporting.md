# ADR 0003: Make reporting evidence-first and AI-optional

- Status: Proposed
- Date: 2026-07-31

## Context

Situation reports must remain available at negligible cost and must be factually traceable. General search/crawling and agent-led reports introduce unpredictable external failures and can invent unsupported statements.

## Decision

Create typed facts, material changes, citations, timeline entries, and structured report sections deterministically. Render a complete template report by default. Offer an optional, operator-triggered model renderer that receives only a bounded evidence bundle and returns validated cited JSON. Humans publish.

## Consequences

- A zero-cost installation has complete reporting capability.
- Citation validation is a domain invariant, not a prompt request.
- AI can improve readability without owning collection, scoring, correlation, or publication.
- More effort is required in source normalization and deterministic fact extraction.

## Rejected alternatives

- Autonomous web-research/report agents: non-reproducible, expensive, and unsafe for core truth.
- Vector RAG by default: datastore/model complexity without a demonstrated retrieval need.
- Model-only summaries: unavailable or degraded when provider credentials/quota fail.
