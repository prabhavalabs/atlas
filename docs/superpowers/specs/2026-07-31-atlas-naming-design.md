# Atlas naming design

- Status: Approved
- Date: 2026-07-31
- Owner: Prabhava Labs

## Decision

The product name is **Atlas**. It is a standalone product name endorsed as **“Atlas by Prabhava Labs”** where organizational context is useful.

## Positioning

Atlas is calm, global, memorable, and geographic. The name supports a product centered on maps, event histories, evidence, and understanding how a disaster changes over time. It avoids alarmist or militaristic language.

The public tagline is:

> Global disaster intelligence, built in the open.

The short description is:

> An open-source platform for monitoring disasters, understanding what changed, and tracing every published fact to its source.

## Naming system

| Surface | Approved form |
|---|---|
| Product | Atlas |
| Endorsed product | Atlas by Prabhava Labs |
| GitHub repository | `prabhavalabs/atlas` |
| Go module | `github.com/prabhavalabs/atlas` |
| CLI/binary | `atlas` |
| Container image | `ghcr.io/prabhavalabs/atlas` |
| Configuration prefix | `ATLAS_` |
| Database/application identifier | `atlas` |

Atlas is not an acronym and must not be written as `ATLAS` except in configuration-variable prefixes. Public copy should pair the name with the disaster-monitoring description when context is not obvious.

## Voice

Atlas communicates with precision, restraint, and visible uncertainty. It uses “verified,” “preliminary,” “estimated,” “conflicting,” and “unknown” deliberately. It does not claim to be an emergency authority and does not write alarmist headlines.

## Release safeguards

The GitHub identity is approved. Before a hosted public release, Prabhava Labs will complete a formal trademark/domain review and select the canonical public domain. Centralized product metadata and design tokens will keep a future legal rename possible without changing the data model or architecture.
