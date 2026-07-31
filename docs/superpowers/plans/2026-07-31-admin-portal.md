# Administration portal implementation plan

> **For the implementing agent:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Give a small trusted editorial team a fast, auditable workflow to resolve ambiguous data, edit cited timelines/reports, and publish or correct public events.

**Architecture:** The `/admin` code-split shell uses secure cookie sessions, CSRF-protected generated mutations, TanStack Query server state, and revision-based concurrency. Every material command requires a reason and appends an immutable audit record.

**Tech stack:** React/Vite/TypeScript, TanStack Query mutations/Router, shadcn/Radix, Go admin API, PostgreSQL audit/review tables, Playwright/axe.

### Task 1: Implement admin shell and session lifecycle

**Files:**

- Create: `web/src/routes/admin/_layout.tsx`
- Create: `web/src/routes/admin/login.tsx`
- Create: `web/src/features/auth/session-provider.tsx`
- Create: `web/src/features/auth/login-form.tsx`
- Create: `web/src/components/layout/admin-shell.tsx`
- Create: `web/src/features/auth/__tests__/session.test.tsx`

**Steps:**

1. Test login success/failure/rate limit, CSRF acquisition, expiry, logout, session rotation, editor/admin navigation, and direct protected-route access.
2. Keep session state in TanStack Query; never store credentials/session tokens in local storage or Zustand.
3. Build dense but accessible sidebar/topbar with active route, skip link, user role, source degradation badge, and logout.
4. On expiry preserve unsaved local form draft and send user to login with a clear recovery path.
5. Verify public bundle does not include admin routes until requested.

### Task 2: Build reasoned mutation and conflict primitives

**Files:**

- Create: `web/src/features/admin/mutations/reason-dialog.tsx`
- Create: `web/src/features/admin/mutations/revision-conflict.tsx`
- Create: `web/src/features/admin/mutations/use-admin-mutation.ts`
- Create: `web/src/features/admin/mutations/__tests__/mutation.test.tsx`

**Steps:**

1. Require non-trivial reason for merge/split/suppress/publish/unpublish/correction/source actions; enforce server-side too.
2. Send current revision/`If-Match`; test 409 shows server changes and lets editor reload/copy local work rather than overwrite.
3. Use optimistic updates only for reversible queue assignment/local labels; publication and evidence changes wait for server success.
4. Map stable problem codes to actionable messages; retain request ID for support.
5. Restore focus and announce success/failure accessibly.

### Task 3: Implement backend review queue

**Files:**

- Create: `migrations/00010_review_items.sql`
- Create: `queries/adminapi/review.sql`
- Create: `internal/adminapi/review.go`
- Create: `internal/adminapi/review_integration_test.go`
- Create: `internal/event/review.go`

**Steps:**

1. Derive/upsert review items for ambiguous correlation, high-priority new event, material change, conflict, report draft, stale source, poison job, and correction flag.
2. Test idempotent reason keys, priority/age ordering, resolve/reopen, target deletion, role visibility, and source recovery.
3. Queue item explains why it exists with safe summary and recommended action; no opaque job terminology.
4. Resolving domain command resolves linked item transactionally; manual dismiss requires reason/audit.
5. Add bounded filters/pagination and counts by reason.

### Task 4: Build review queue UI

**Files:**

- Create: `web/src/routes/admin/index.tsx`
- Create: `web/src/features/review/review-queue.tsx`
- Create: `web/src/features/review/review-item.tsx`
- Create: `web/src/features/review/__tests__/review-queue.test.tsx`

**Steps:**

1. Test normal/empty/stale/failure, grouped reasons, filter/deep link, pagination, resolve/dismiss conflict, editor/admin permissions, and keyboard navigation.
2. Display reason, safe target summary, change time, confidence/priority, source status, and primary action.
3. No bulk publishing. Batch dismiss is deferred until measured need and safety design.
4. Preserve queue place when returning from event workspace.
5. Add live count updates without moving focused item unexpectedly.

### Task 5: Implement event workspace read/command APIs

**Files:**

- Create: `internal/adminapi/workspace.go`
- Create: `internal/adminapi/events.go`
- Create: `internal/adminapi/evidence.go`
- Create: `internal/adminapi/workspace_integration_test.go`

**Steps:**

1. Return bounded workspace DTO: canonical revision, evidence/change history, correlation explanation, geometry, timeline drafts/publications, report revisions, publication validation, and related review items.
2. Add revision-checked commands for canonical fields, link/unlink/suppress evidence, geometry visibility, lifecycle/verification, and manual source URL metadata.
3. Test role checks, reason/audit, citation invalidation, public isolation, and transaction rollback on every command failure.
4. Suppression preserves observation/history and blocks dependent draft publication; published content requires correction workflow.
5. Manual URLs pass SSRF/rights policy and do not trigger browser crawling.

### Task 6: Build event workspace UI

**Files:**

- Create: `web/src/routes/admin/events/$id.tsx`
- Create: `web/src/features/admin-event/workspace.tsx`
- Create: `web/src/features/admin-event/evidence-panel.tsx`
- Create: `web/src/features/admin-event/event-form.tsx`
- Create: `web/src/features/admin-event/publication-panel.tsx`
- Create: `web/src/features/admin-event/__tests__/workspace.test.tsx`

**Steps:**

1. Build approved three-part desktop/stacked mobile layout with resizable panels stored only as UI preference.
2. Test evidence filter/diff/link/suppress, field edit validation, geometry source/precision, concurrent revision, unsaved navigation, and role restrictions.
3. Show source values beside canonical value; unknown/conflict/predicted labels remain explicit.
4. Publication panel lists blocking errors, preview link, current public revision, and correction requirement.
5. Every evidence item opens exact original source/version metadata without executing source HTML.

### Task 7: Build correlation merge/split review

**Files:**

- Create: `web/src/features/correlation/candidate-review.tsx`
- Create: `web/src/features/correlation/merge-dialog.tsx`
- Create: `web/src/features/correlation/split-dialog.tsx`
- Create: `web/src/features/correlation/__tests__/correlation.test.tsx`

**Steps:**

1. Present side-by-side map/time/type/title/source values and explain individual score features/threshold band.
2. Test merge, keep separate, split selected observations, undo unpublished merge, revision conflict, and keyboard flow.
3. Preview affected aliases/timeline/reports/public slug before confirmation.
4. Require reason and show immutable result/audit link.
5. Never label a probabilistic score as certainty.

### Task 8: Build timeline editor

**Files:**

- Create: `web/src/features/admin-timeline/timeline-editor.tsx`
- Create: `web/src/features/admin-timeline/timeline-form.tsx`
- Create: `web/src/features/admin-timeline/citation-picker.tsx`
- Create: `web/src/features/admin-timeline/__tests__/timeline-editor.test.tsx`

**Steps:**

1. Test candidate accept/edit/reject, manual entry, occurred/published time, materiality/category, citation add/remove, conflict, correction, and preview.
2. Server blocks factual publication with zero citations; UI explains exact missing support.
3. Citation picker searches only event-linked evidence and displays source/version/time/field support.
4. Preserve revision history and show public/draft distinction.
5. Reject arbitrary rich HTML; use constrained text/links only.

### Task 9: Build structured report editor and optional generation

**Files:**

- Create: `internal/adminapi/reports.go`
- Create: `internal/adminapi/reports_integration_test.go`
- Create: `web/src/features/admin-report/report-editor.tsx`
- Create: `web/src/features/admin-report/evidence-inspector.tsx`
- Create: `web/src/features/admin-report/generate-section.tsx`
- Create: `web/src/features/admin-report/__tests__/report-editor.test.tsx`

**Steps:**

1. Test template draft, section edit, citation change, regenerate one section, preserve hand-edited sections, invalid model output fallback, budget exceeded, concurrent edit, preview, publish/correct.
2. Display paragraph-level supporting evidence and unknown/conflict validation.
3. Generation dialog shows provider/model, evidence count, estimated budget, privacy note, and explicit action; no background generation.
4. Label generated draft until human review; store generation method in publication metadata.
5. Publishing uses current evidence hash/revision and fails safely if evidence changed since draft.

### Task 10: Build sources, jobs, and system health administration

**Files:**

- Create: `web/src/routes/admin/sources.tsx`
- Create: `web/src/routes/admin/health.tsx`
- Create: `web/src/features/admin-source/source-table.tsx`
- Create: `web/src/features/admin-source/run-history.tsx`
- Create: `web/src/features/admin-health/job-table.tsx`
- Create: `web/src/features/admin-health/__tests__/health.test.tsx`

**Steps:**

1. Test healthy/stale/circuit-open/never-run/partial source, run now, reset breaker, enable/disable, retry poison job, and role denial.
2. Translate health into freshness, last success/change, error streak/class, next retry, counts, and action; redact secrets/raw bodies.
3. Require reason/audit for state/config changes; run-now is idempotent and rate constrained.
4. Show database/storage/backup/schema/app version and restore-test age with thresholds.
5. No arbitrary job payload editor or shell command UI.

### Task 11: Build audit and correction flows

**Files:**

- Create: `web/src/routes/admin/audit.tsx`
- Create: `web/src/features/audit/audit-list.tsx`
- Create: `web/src/features/publication/correction-dialog.tsx`
- Create: `web/src/features/audit/__tests__/audit.test.tsx`

**Steps:**

1. Test audit filters by actor/action/target/time, bounded before/after detail, pagination, and editor/admin visibility.
2. Correction requires public explanation, creates new revision, preserves old URL/history, and updates dependent review item.
3. Unpublish/withdraw displays public correction/withdrawal state rather than silent 404 when safety/legal policy allows.
4. Audit data is immutable from application UI/API.
5. Export is a bounded JSON/CSV admin operation with rate/size limits and audit entry.

### Task 12: Admin verification checkpoint

**Files:**

- Create: `web/e2e/admin-review-publish.spec.ts`
- Create: `web/e2e/admin-correction.spec.ts`
- Create: `web/e2e/admin-security.spec.ts`
- Create: `docs/release-checklists/admin-portal.md`

**Steps:**

1. E2E: login, resolve correlation, edit cited timeline/report, publish, inspect public, correct, inspect revision/audit.
2. E2E/session security: CSRF, expiry, role denial, rate limit, concurrent revision, XSS fixture, SSRF URL rejection.
3. Run axe, keyboard at 200%, mobile stacked workspace, reduced motion, and screen-reader smoke.
4. Verify no admin draft/raw/audit/session response is cacheable or exposed through public API.
5. Demonstrate all material actions have reason/audit and rollback/correction path.
