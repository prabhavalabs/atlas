# ADR 0006: Use Apache-2.0 for the project

- Status: Accepted
- Date: 2026-07-31

## Context

The platform should be genuinely open-source and easy for governments, NGOs, researchers, companies, and volunteers to deploy and contribute to. Reference repositories have MIT and AGPL licenses, and existing components need explicit license review.

## Decision

Use Apache License 2.0 for original backend/frontend code because it is permissive, OSI-approved, and includes an explicit patent grant. Maintain `LICENSE`, `NOTICE`, dependency notices, dataset attribution, and a machine-readable SBOM. Do not copy AGPL code into the project.

## Consequences

- Broad reuse and contribution are easier.
- Closed hosted forks are permitted; the community does not receive network-copyleft protection.
- MIT/BSD code can generally be reused with notices.
- Every reused internal component must first have confirmed ownership and a compatible explicit license.

## Rejected alternative

AGPL-3.0-only would ensure that modified network deployments publish source, but it would reduce permissive adoption and complicate use with organizations that standardize on permissive licenses. Relicensing later would require contributor consent.
