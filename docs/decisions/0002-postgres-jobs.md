# ADR 0002: Use PostgreSQL-backed durable jobs

- Status: Proposed
- Date: 2026-07-31

## Context

Source polling and event/report processing need retries, idempotency, delayed execution, and visibility. An external queue or workflow engine would add credentials, persistence, deployment, monitoring, and consistency boundaries.

## Decision

Use a `jobs` table claimed with `FOR UPDATE SKIP LOCKED`, leases/heartbeats, idempotency keys, bounded retries, and dead-job visibility. Create jobs in the same transaction as their triggering domain change. Use advisory locks for one active scheduler leader.

## Consequences

- No Redis/broker/workflow service in the core deployment.
- Atomic domain-and-job state prevents missing webhook-style work.
- PostgreSQL load must be observed; handlers must be short and idempotent.
- A future broker is possible behind the job interface if measured throughput requires it.

## Rejected alternatives

- Redis queue: another volatile/durable service and non-atomic database handoff.
- Cloud queue/functions: external cost and deployment dependency.
- In-memory scheduler only: loses work on restart and cannot provide audit/replay.
