package adminapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAttemptLimiterCapsFailuresWithinWindowAndRecovers(t *testing.T) {
	limiter := newAttemptLimiter(2, time.Minute)
	now := time.Date(2026, time.July, 31, 20, 0, 0, 0, time.UTC)

	require.True(t, limiter.Allow("client|user@example.org", now))
	require.True(t, limiter.Allow("client|user@example.org", now.Add(time.Second)))
	require.False(t, limiter.Allow("client|user@example.org", now.Add(2*time.Second)))
	require.True(t, limiter.Allow("client|other@example.org", now.Add(2*time.Second)))
	require.True(t, limiter.Allow("client|user@example.org", now.Add(61*time.Second)))
}

func TestAttemptLimiterResetRemovesFailureHistory(t *testing.T) {
	limiter := newAttemptLimiter(1, time.Minute)
	now := time.Date(2026, time.July, 31, 20, 0, 0, 0, time.UTC)

	require.True(t, limiter.Allow("key", now))
	require.False(t, limiter.Allow("key", now))
	limiter.Reset("key")
	require.True(t, limiter.Allow("key", now))
}
