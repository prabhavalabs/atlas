# Atlas interface direction

Atlas uses a calm cartographic newsroom aesthetic: information-dense, composed,
and easy to scan during a fast-moving incident. The interface should feel like a
trusted public utility rather than a command center or a consumer news feed.

The four approved foundation concepts in this directory are the visual source of
truth for the first release:

- `public-overview-desktop.png`
- `public-overview-mobile.png`
- `admin-login-desktop.png`
- `admin-review-desktop.png`

## Experience principles

1. Lead with verified facts, timestamps, and provenance.
2. Keep the event list usable without a map; the map is a spatial aid, not the
   only navigation surface.
3. Reserve hazard colours for incident meaning. Navigation and controls use ink,
   grey, and teal.
4. Prefer open layouts, thin rules, and quiet selected states over stacked cards
   and heavy shadows.
5. Preserve a clear distinction between observed facts, human-authored summaries,
   and machine-assisted drafts.
6. Make every core task keyboard accessible and understandable at 200% zoom.

## Visual tokens

The CSS implementation uses OKLCH values and semantic token names. These values
are starting points; contrast requirements take precedence over visual matching.

| Token | Value | Purpose |
| --- | --- | --- |
| `background` | `oklch(1 0 0)` | True-white canvas |
| `foreground` | `oklch(0.20 0.035 250)` | Deep ink text |
| `muted-foreground` | `oklch(0.48 0.02 250)` | Secondary copy |
| `border` | `oklch(0.90 0.015 240)` | Rules and boundaries |
| `primary` | `oklch(0.53 0.11 195)` | Atlas teal |
| `accent` | `oklch(0.95 0.025 195)` | Selected and hover state |
| `destructive` | `oklch(0.57 0.22 27)` | Destructive action only |
| `hazard-severe` | `oklch(0.60 0.20 28)` | Severe incident data |
| `hazard-moderate` | `oklch(0.70 0.16 55)` | Moderate incident data |
| `hazard-watch` | `oklch(0.78 0.15 88)` | Watch-level incident data |

- Typeface: Geist Variable.
- Base text: 16px on public pages; dense administrative metadata may use 13–14px.
- Radius: 8px for controls and compact surfaces.
- Elevation: none by default; a restrained shadow is allowed only for transient
  overlays such as menus and sheets.
- Motion: 100–200ms for direct manipulation; respect `prefers-reduced-motion`.

## Layout and behaviour

### Public overview

- Desktop uses a persistent incident list beside the map and selected-event
  details. Filters stay close to the list.
- Mobile is list-first. A labelled **View map** control opens the spatial view in
  a sheet so the core content remains useful on constrained devices.
- Every map marker has an equivalent event row. Selecting either updates the same
  local UI state and never changes verification data.
- Loading, empty, stale, partial-source, and unavailable-map states must be
  explicit. A failed map must not hide the event list.
- Event detail pages present the concise situation summary, key facts, evidence,
  and a chronological timeline in that order.

### Administration

- The sign-in screen is deliberately quiet and only requests email and password.
- The review queue is table-first, with status and confidence conveyed by text as
  well as colour.
- Editing and publication controls keep the current revision visible and preserve
  authorship, evidence, and timestamps in the audit trail.
- Potentially destructive or public-facing actions require explicit labels and a
  confirmation step where reversal would be costly.

## Accessibility baseline

- WCAG 2.2 AA is the minimum target.
- Use semantic landmarks, a skip link, visible focus, and correctly associated
  labels and descriptions.
- Do not use colour as the sole indicator of hazard level, confidence, or status.
- Interactive targets are at least 44 by 44 CSS pixels on touch layouts.
- Tables remain navigable by assistive technology and become labelled stacked
  rows on narrow screens rather than horizontal-scroll traps.
- Dates expose an absolute UTC value; relative time is supplementary.

## Content voice

Atlas copy is factual, concise, and transparent about uncertainty. Avoid urgency
language unless an authoritative source has issued it. Prefer “reported”,
“confirmed”, and “last checked” over vague confidence claims. Machine-assisted
text is always a draft until an administrator publishes it.
