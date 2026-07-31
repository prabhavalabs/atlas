// Package jobs provides durable, idempotent work leases backed by PostgreSQL.
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// State is the persisted job lifecycle.
type State string

const (
	// StateQueued is durable work waiting to be leased.
	StateQueued State = "queued"
	// StateRunning is work currently owned by a live lease.
	StateRunning State = "running"
	// StateSucceeded is work that completed successfully.
	StateSucceeded State = "succeeded"
	// StateFailed is work quarantined after its final attempt.
	StateFailed State = "failed"
)

var (
	// ErrNoJob means no queued work is currently available.
	ErrNoJob = errors.New("no job available")
	// ErrLeaseLost means a worker no longer owns the running job.
	ErrLeaseLost = errors.New("job lease no longer owned")
)

// Job is a durable unit of asynchronous work.
type Job struct {
	ID               uuid.UUID
	Type             string
	IdempotencyKey   string
	State            State
	Payload          json.RawMessage
	Attempt          int
	MaxAttempts      int
	AvailableAt      time.Time
	LeaseOwner       *string
	LeaseExpiresAt   *time.Time
	LastErrorCode    string
	LastErrorMessage string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time
}

// EnqueueInput defines unique work. Re-enqueueing the same type and key is safe.
type EnqueueInput struct {
	Type           string
	IdempotencyKey string
	Payload        json.RawMessage
	MaxAttempts    int
	AvailableAt    time.Time
}

// FailureInput records a bounded retry or terminal failure.
type FailureInput struct {
	ID      uuid.UUID
	Owner   string
	Code    string
	Message string
	Now     time.Time
	RetryAt time.Time
}

// Repository persists jobs and lease transitions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed job repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Enqueue creates work once per type and idempotency key.
func (repository *Repository) Enqueue(ctx context.Context, input EnqueueInput) (Job, error) {
	input.Type = strings.TrimSpace(input.Type)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.Type == "" || len(input.Type) > 100 || input.IdempotencyKey == "" || len(input.IdempotencyKey) > 500 {
		return Job{}, errors.New("job type and idempotency key are required and bounded")
	}
	if len(input.Payload) == 0 {
		input.Payload = json.RawMessage(`{}`)
	}
	if !json.Valid(input.Payload) {
		return Job{}, errors.New("job payload must be valid JSON")
	}
	if input.MaxAttempts == 0 {
		input.MaxAttempts = 5
	}
	if input.MaxAttempts < 1 || input.MaxAttempts > 100 {
		return Job{}, errors.New("job max attempts must be between 1 and 100")
	}
	if input.AvailableAt.IsZero() {
		input.AvailableAt = time.Now().UTC()
	}

	row := repository.pool.QueryRow(ctx, `
		INSERT INTO jobs (id, type, idempotency_key, payload, max_attempts, available_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (type, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id, type, idempotency_key, state, payload, attempt, max_attempts,
			available_at, lease_owner, lease_expires_at,
			COALESCE(last_error_code, ''), COALESCE(last_error_message, ''),
			created_at, updated_at, completed_at
	`, uuid.New(), input.Type, input.IdempotencyKey, input.Payload, input.MaxAttempts, input.AvailableAt)
	job, err := scanJob(row)
	if err != nil {
		return Job{}, fmt.Errorf("enqueue job: %w", err)
	}
	return job, nil
}

// Claim atomically leases the oldest available job. Expired leases are first recovered.
func (repository *Repository) Claim(ctx context.Context, owner string, now time.Time, lease time.Duration) (Job, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || lease <= 0 {
		return Job{}, errors.New("job owner and positive lease are required")
	}
	transaction, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Job{}, fmt.Errorf("begin job claim: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	_, err = transaction.Exec(ctx, `
		UPDATE jobs
		SET state = 'queued', lease_owner = NULL, lease_expires_at = NULL,
			available_at = $1, updated_at = $1,
			last_error_code = 'lease_expired', last_error_message = 'worker lease expired'
		WHERE state = 'running' AND lease_expires_at <= $1
	`, now)
	if err != nil {
		return Job{}, fmt.Errorf("recover expired job leases: %w", err)
	}

	row := transaction.QueryRow(ctx, `
		WITH next_job AS (
			SELECT id FROM jobs
			WHERE state = 'queued' AND available_at <= $1
			ORDER BY available_at, created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE jobs j
		SET state = 'running', attempt = j.attempt + 1, lease_owner = $2,
			lease_expires_at = $3, updated_at = $1
		FROM next_job
		WHERE j.id = next_job.id
		RETURNING j.id, j.type, j.idempotency_key, j.state, j.payload,
			j.attempt, j.max_attempts, j.available_at, j.lease_owner,
			j.lease_expires_at, COALESCE(j.last_error_code, ''),
			COALESCE(j.last_error_message, ''), j.created_at, j.updated_at,
			j.completed_at
	`, now, owner, now.Add(lease))
	job, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrNoJob
	}
	if err != nil {
		return Job{}, fmt.Errorf("claim job: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return Job{}, fmt.Errorf("commit job claim: %w", err)
	}
	return job, nil
}

// Succeed completes a job only while owner still holds its lease.
func (repository *Repository) Succeed(ctx context.Context, id uuid.UUID, owner string, now time.Time) error {
	result, err := repository.pool.Exec(ctx, `
		UPDATE jobs
		SET state = 'succeeded', lease_owner = NULL, lease_expires_at = NULL,
			completed_at = $3, updated_at = $3,
			last_error_code = NULL, last_error_message = NULL
		WHERE id = $1 AND state = 'running' AND lease_owner = $2
	`, id, owner, now)
	if err != nil {
		return fmt.Errorf("complete job: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrLeaseLost
	}
	return nil
}

// Fail records a safe error and either schedules a retry or quarantines the job.
func (repository *Repository) Fail(ctx context.Context, input FailureInput) (State, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Message = strings.TrimSpace(input.Message)
	if input.Owner == "" || input.Code == "" || len(input.Code) > 100 || len(input.Message) > 1000 {
		return "", errors.New("job failure owner, bounded code, and bounded message are required")
	}
	var state State
	err := repository.pool.QueryRow(ctx, `
		UPDATE jobs
		SET state = CASE WHEN attempt >= max_attempts THEN 'failed' ELSE 'queued' END,
			available_at = CASE WHEN attempt >= max_attempts THEN available_at ELSE $6 END,
			lease_owner = NULL, lease_expires_at = NULL,
			last_error_code = $3, last_error_message = $4, updated_at = $5::timestamptz,
			completed_at = CASE WHEN attempt >= max_attempts THEN $5::timestamptz ELSE NULL::timestamptz END
		WHERE id = $1 AND state = 'running' AND lease_owner = $2
		RETURNING state
	`, input.ID, input.Owner, input.Code, input.Message, input.Now, input.RetryAt).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrLeaseLost
	}
	if err != nil {
		return "", fmt.Errorf("fail job: %w", err)
	}
	return state, nil
}

// Get returns one job for administration and tests.
func (repository *Repository) Get(ctx context.Context, id uuid.UUID) (Job, error) {
	job, err := scanJob(repository.pool.QueryRow(ctx, `
		SELECT id, type, idempotency_key, state, payload, attempt, max_attempts,
			available_at, lease_owner, lease_expires_at,
			COALESCE(last_error_code, ''), COALESCE(last_error_message, ''),
			created_at, updated_at, completed_at
		FROM jobs WHERE id = $1
	`, id))
	if err != nil {
		return Job{}, fmt.Errorf("get job: %w", err)
	}
	return job, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(row scanner) (Job, error) {
	var job Job
	err := row.Scan(
		&job.ID, &job.Type, &job.IdempotencyKey, &job.State, &job.Payload,
		&job.Attempt, &job.MaxAttempts, &job.AvailableAt, &job.LeaseOwner,
		&job.LeaseExpiresAt, &job.LastErrorCode, &job.LastErrorMessage,
		&job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
	)
	return job, err
}
