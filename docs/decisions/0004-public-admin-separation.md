# ADR 0004: Separate public and admin experiences

- Status: Proposed
- Date: 2026-07-31

## Context

Public readers need immediate account-free access and a calm narrative. Editors need dense evidence, mutations, health, and audit controls. Combining both creates authentication gates and confusing navigation.

## Decision

Use one React/Vite application workspace with separate public and `/admin` route shells, layouts, code-split bundles, API clients, and authorization boundaries. Public APIs are read-only and cacheable. Admin APIs require local secure sessions and CSRF protection.

## Consequences

- Public monitoring never redirects to login.
- Admin dependencies and dense UI need not burden the initial public bundle.
- Shared primitives/design tokens remain reusable.
- Route and API authorization are tested independently.

## Rejected alternatives

- One dashboard with role-hidden controls: public IA and bundle remain polluted by admin concerns.
- Two repositories/deployments: duplicates tooling and release coordination without security benefit at MVP scale.
- Next.js: explicitly outside the desired stack and unnecessary for the cacheable Vite SPA/API architecture.
