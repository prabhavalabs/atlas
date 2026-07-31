# Contributing to Atlas

Thank you for helping build trustworthy, accessible disaster information in the open.

## Before contributing

1. Read the [product vision](docs/01-product-vision.md), [architecture](docs/05-architecture.md), and relevant ADRs.
2. Search existing issues and discussions before proposing overlapping work.
3. Use an issue for material product, source, data-model, or architecture changes before implementation.
4. Keep the initial complexity budget: one Go application, one PostGIS database, one frontend workspace, and three core production containers.

## Source proposals

A new source proposal must include:

- owning authority and geographic/hazard scope;
- official access and terms/license URLs;
- attribution and permitted retention;
- authentication, rate guidance, and expected cadence;
- stable identity, update, cancellation, and test/exercise semantics;
- sample fixtures that may legally be stored in the repository;
- expected failure, stale, and recovery behavior.

Do not submit full copyrighted news articles, secrets, private documents, or fixtures containing unnecessary personal data.

## Pull requests

- Keep each pull request focused on one logical change.
- Add or update tests before changing behavior.
- Preserve citations, provenance, and unknown/conflict semantics.
- Update OpenAPI, migrations, ADRs, operations, costs, and public methodology when affected.
- Use Conventional Commits such as `docs:`, `feat:`, `fix:`, `test:`, or `chore:`.
- Never bypass hooks or checks to make a change pass.

The detailed phase plans under `docs/superpowers/plans/` define the expected test-first sequence for initial development.

## Review standards

Changes are reviewed for factual safety, accessibility, source rights, reproducibility, failure behavior, operational cost, and maintainability by a small team. A useful feature that cannot degrade safely or be run by a solo maintainer is not ready for inclusion.

By submitting a contribution, you agree that it is provided under the repository's Apache-2.0 license and that you have the right to contribute it.
