//go:build integration

package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/prabhavalabs/atlas/internal/jobs"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/stretchr/testify/require"
)

func TestRepositoryEnqueuesIdempotentlyAndClaimsOneLease(t *testing.T) {
	repository := setupJobs(t)
	ctx := context.Background()
	now := time.Date(2026, time.July, 31, 20, 0, 0, 0, time.UTC)

	first, err := repository.Enqueue(ctx, jobs.EnqueueInput{
		Type: "source.fetch", IdempotencyKey: "fixture:2026-07-31T20", Payload: json.RawMessage(`{"source":"fixture"}`), AvailableAt: now,
	})
	require.NoError(t, err)
	second, err := repository.Enqueue(ctx, jobs.EnqueueInput{
		Type: "source.fetch", IdempotencyKey: "fixture:2026-07-31T20", Payload: json.RawMessage(`{"source":"fixture"}`), AvailableAt: now,
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)

	claimed, err := repository.Claim(ctx, "worker-a", now, time.Minute)
	require.NoError(t, err)
	require.Equal(t, first.ID, claimed.ID)
	require.Equal(t, jobs.StateRunning, claimed.State)
	require.Equal(t, 1, claimed.Attempt)

	_, err = repository.Claim(ctx, "worker-b", now, time.Minute)
	require.ErrorIs(t, err, jobs.ErrNoJob)
	require.NoError(t, repository.Succeed(ctx, claimed.ID, "worker-a", now.Add(time.Second)))
}

func TestRepositoryRetriesAndEventuallyQuarantinesFailedJob(t *testing.T) {
	repository := setupJobs(t)
	ctx := context.Background()
	now := time.Date(2026, time.July, 31, 20, 0, 0, 0, time.UTC)
	enqueued, err := repository.Enqueue(ctx, jobs.EnqueueInput{
		Type: "source.fetch", IdempotencyKey: "failure", Payload: json.RawMessage(`{}`), MaxAttempts: 2, AvailableAt: now,
	})
	require.NoError(t, err)

	first, err := repository.Claim(ctx, "worker-a", now, time.Minute)
	require.NoError(t, err)
	state, err := repository.Fail(ctx, jobs.FailureInput{
		ID: first.ID, Owner: "worker-a", Code: "upstream_timeout", Message: "source timed out", Now: now.Add(time.Second), RetryAt: now.Add(time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, jobs.StateQueued, state)
	_, err = repository.Claim(ctx, "worker-a", now.Add(30*time.Second), time.Minute)
	require.ErrorIs(t, err, jobs.ErrNoJob)

	second, err := repository.Claim(ctx, "worker-a", now.Add(time.Minute), time.Minute)
	require.NoError(t, err)
	state, err = repository.Fail(ctx, jobs.FailureInput{
		ID: second.ID, Owner: "worker-a", Code: "invalid_payload", Message: "source returned invalid data", Now: now.Add(time.Minute + time.Second), RetryAt: now.Add(2 * time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, jobs.StateFailed, state)
	_, err = repository.Claim(ctx, "worker-a", now.Add(3*time.Minute), time.Minute)
	require.True(t, errors.Is(err, jobs.ErrNoJob))

	stored, err := repository.Get(ctx, enqueued.ID)
	require.NoError(t, err)
	require.Equal(t, jobs.StateFailed, stored.State)
	require.Equal(t, "invalid_payload", stored.LastErrorCode)
}

func setupJobs(t *testing.T) *jobs.Repository {
	t.Helper()
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))
	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE jobs")
	require.NoError(t, err)
	return jobs.NewRepository(pool)
}
