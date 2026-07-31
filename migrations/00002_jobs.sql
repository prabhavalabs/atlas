-- +goose Up
CREATE TABLE jobs (
    id uuid PRIMARY KEY,
    type text NOT NULL,
    idempotency_key text NOT NULL,
    state text NOT NULL DEFAULT 'queued' CHECK (state IN ('queued', 'running', 'succeeded', 'failed')),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    max_attempts integer NOT NULL DEFAULT 5 CHECK (max_attempts > 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_expires_at timestamptz,
    last_error_code text,
    last_error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    UNIQUE (type, idempotency_key)
);

CREATE INDEX jobs_claim_idx ON jobs (available_at, created_at)
WHERE state = 'queued';

CREATE INDEX jobs_expired_lease_idx ON jobs (lease_expires_at)
WHERE state = 'running';

-- +goose Down
DROP TABLE IF EXISTS jobs;
