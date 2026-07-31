# Public web application implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Deliver an account-free, responsive, accessible world overview and cited event narrative with synchronized list/map, search, filters, timelines, sources, and honest degraded states.

**Architecture:** A code-split React/Vite public shell uses generated OpenAPI clients and TanStack Query for server state. URL search parameters own shareable filter state; Zustand owns only ephemeral map/layout preferences. MapLibre renders event geometry with a semantic list/text alternative.

**Tech stack:** React, Vite, TypeScript, TanStack Router/Query, Zustand, shadcn/ui, Tailwind, MapLibre GL JS, PMTiles, MSW, Vitest/Testing Library/axe, Playwright.

### Task 1: Complete the public design system and state matrix

**Files:**

- Create: `web/src/styles/tokens.css`
- Create: `web/src/components/ui/`
- Create: `web/src/components/layout/public-shell.tsx`
- Create: `web/src/components/status/`
- Create: `web/src/stories/public-states.tsx`
- Create: `docs/design/public-ui.md`

**Steps:**

1. Confirm identity decision and implement approved typography, color, spacing, radius, elevation, motion, hazard/status semantics, dark theme, and map palettes as CSS variables.
2. Build actual shadcn/Radix primitives with visible focus and no color-only state. Do not create custom SVG/icon approximations; use the selected icon library.
3. Render overview/event states: loading, cached, empty, stale, partial outage, full source outage, error, narrow mobile, 200% zoom, dark, reduced motion.
4. Test keyboard order and axe for every state matrix entry before feature wiring.
5. Document token and content rules so hazard colors remain data-only.

### Task 2: Implement URL-owned filters and query DTOs

**Files:**

- Create: `web/src/features/events/filter-schema.ts`
- Create: `web/src/features/events/use-event-filters.ts`
- Create: `web/src/features/events/event-filters.tsx`
- Create: `web/src/features/events/__tests__/filters.test.tsx`

**Steps:**

1. Define/parse canonical URL params for type, priority, lifecycle, time, region, search, viewport opt-in, sort, and page; invalid values fall back safely.
2. Test back/forward, deep link, reset, mobile drawer, keyboard controls, and generated API parameter mapping.
3. Debounce text search while committing discrete filters immediately; keep list/map on one query key.
4. Announce result changes accessibly without stealing focus.
5. Ensure no precise user location is persisted/logged by default.

### Task 3: Build global overview list and coverage status

**Files:**

- Create: `web/src/routes/index.tsx`
- Create: `web/src/features/events/event-list.tsx`
- Create: `web/src/features/events/event-card.tsx`
- Create: `web/src/features/coverage/coverage-strip.tsx`
- Create: `web/src/features/events/__tests__/event-list.test.tsx`

**Steps:**

1. Write MSW tests for normal, empty-filter, first-install-no-coverage, stale, partial failure, total failure with cached data, pagination, and revalidation 304.
2. Render semantic heading/list; card includes type, title, place, lifecycle, source severity, public priority, confidence, material change, verified time, and source count.
3. Add textual/tooltip explanations for every status; distinguish “not monitored,” “no matching event,” and “source stale.”
4. Preserve cached list during background failure and expose retry/source methodology.
5. Enforce skeleton layout stability and no sign-in request/redirect on this route.

### Task 4: Build accessible MapLibre overview

**Files:**

- Create: `web/src/features/map/event-map.tsx`
- Create: `web/src/features/map/map-style.ts`
- Create: `web/src/features/map/map-store.ts`
- Create: `web/src/features/map/map-fallback.tsx`
- Create: `web/src/features/map/__tests__/event-map.test.tsx`

**Steps:**

1. Lazy-load map bundle only when map view is visible. Configure style URL, PMTiles protocol, attribution, max bounds, cooperative gestures, and reduced-motion behavior.
2. Render points/clusters and approved geometry types from API GeoJSON; selected list/event state is synchronized without duplicating server data in Zustand.
3. Test keyboard-accessible map/list switch, no keyboard trap, focus preservation, attribution visibility, map WebGL failure, tile failure, and local low-zoom fallback.
4. Provide a text geometry summary and “view as list” equivalent for every map state.
5. On mobile default to list; map opens full-screen and restores selected item/scroll on close.

### Task 5: Implement event narrative route

**Files:**

- Create: `web/src/routes/events/$slug.tsx`
- Create: `web/src/features/event-detail/event-header.tsx`
- Create: `web/src/features/event-detail/known-facts.tsx`
- Create: `web/src/features/event-detail/situation-report.tsx`
- Create: `web/src/features/event-detail/event-geometry.tsx`
- Create: `web/src/features/event-detail/__tests__/event-page.test.tsx`

**Steps:**

1. Test published event, stale data, preliminary/automatic labels, approximate geometry, conflict, unknown impact, correction, withdrawn/not-found, and previous slug redirect.
2. Build one anchored narrative in the exact order documented in `docs/02-experience-design.md`; no public tabs.
3. Render inline citation links with source/title/time and accessible return-to-claim behavior.
4. Conflicting figures display side by side; unknown sections use explicit wording; estimated exposure is distinct from observed impact.
5. Add metadata/share URL and machine-readable timestamps; defer full SEO prerendering until evidence shows SPA indexing is insufficient.

### Task 6: Implement timeline, sources, and report history

**Files:**

- Create: `web/src/features/timeline/timeline.tsx`
- Create: `web/src/features/timeline/timeline-entry.tsx`
- Create: `web/src/features/sources/source-list.tsx`
- Create: `web/src/features/reports/revision-history.tsx`
- Create: `web/src/features/timeline/__tests__/timeline.test.tsx`

**Steps:**

1. Test chronological groups, dense-day collapse, occurred versus published time, conflict/correction/withdrawal, citation expansion, pagination, and keyboard navigation.
2. Default to material public entries; offer expanded source history without loading it into initial payload.
3. Source list shows authority class, attribution, last successful check, original link, and limitation/freshness.
4. Report history shows as-of/published times, generation method label, editor review status, and correction notes without exposing admin identity unnecessarily.
5. Use semantic lists/time elements/details; no visual-only connector that obscures reading order.

### Task 7: Implement public search

**Files:**

- Create: `web/src/routes/search.tsx`
- Create: `web/src/features/search/search-box.tsx`
- Create: `web/src/features/search/search-results.tsx`
- Create: `web/src/features/search/__tests__/search.test.tsx`

**Steps:**

1. Test title/place/alias/report/article matches, typo/trigram, no result, invalid query, loading/cached/failure, pagination, and safe highlight rendering.
2. Label result kind and matched field; result links restore/share query.
3. Support keyboard shortcut only when it does not conflict with input/assistive technology; document it visibly.
4. Track privacy-preserving aggregate success/no-result counts only if analytics is explicitly enabled.
5. Do not add semantic/model search in this task.

### Task 8: Add methodology, coverage, accessibility, and corrections pages

**Files:**

- Create: `web/src/routes/about.tsx`
- Create: `web/src/routes/methodology.tsx`
- Create: `web/src/routes/coverage.tsx`
- Create: `web/src/routes/accessibility.tsx`
- Create: `web/src/routes/corrections.tsx`
- Create: `web/src/features/content/content-page.test.tsx`

**Steps:**

1. Present generated source registry/attribution, freshness definitions, severity/priority/confidence distinctions, AI policy, limitations, and emergency disclaimer.
2. Coverage page distinguishes enabled source scopes and stale/unavailable regions/hazards.
3. Accessibility statement lists target, known issues, contact, and review date.
4. Corrections page explains public revision history and provides a configured contact channel without adding public accounts.
5. Link pages from header/footer and contextually from stale/estimate/correction UI.

### Task 9: Optimize public bundle and caching

**Files:**

- Modify: `web/vite.config.ts`
- Modify: `web/src/router.tsx`
- Create: `web/scripts/check-bundle.mjs`
- Modify: `deploy/Caddyfile`

**Steps:**

1. Measure baseline; set approved JS/CSS budgets. Split MapLibre, admin, and report-history heavy paths from overview critical path.
2. Use immutable hashes for assets, revalidation for HTML/API, and no caching of admin/session responses.
3. Test slow network/offline navigation to already viewed event with cached query data and visible stale state.
4. Prevent broad service-worker caching in MVP unless the PWA decision is explicitly made.
5. Verify source/map external domains in CSP are deployment-configured and minimal.

### Task 10: Public end-to-end and accessibility verification

**Files:**

- Create: `web/e2e/public-overview.spec.ts`
- Create: `web/e2e/public-event.spec.ts`
- Create: `web/e2e/public-degraded.spec.ts`
- Create: `docs/release-checklists/public-web.md`

**Steps:**

1. Playwright covers filter/share/back, list/map sync, event citations/timeline/revision, search, stale/outage/map failure, and mobile list-first flows.
2. Run axe at route states, keyboard-only at 200% zoom, reduced motion, high contrast, and manual screen-reader smoke on overview/event.
3. Run visual regression for approved desktop/mobile light/dark and degraded states.
4. Measure LCP/INP/CLS and bundle/API budgets on core VPS/network profile.
5. Verify network log contains no auth request and no login redirect for any public journey.
