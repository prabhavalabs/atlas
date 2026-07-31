//go:build integration

package event_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/prabhavalabs/atlas/internal/event"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/prabhavalabs/atlas/internal/source/fixture"
	"github.com/stretchr/testify/require"
)

func TestFixtureImportPublishesOneIdempotentPublicEvent(t *testing.T) {
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL, "ATLAS_TEST_DATABASE_URL must reference a disposable database")
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))

	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE timeline_entries, event_observations, event_revisions, events, observations, sources RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	fixturePath := filepath.Join("..", "..", "testdata", "sources", "fixture", "minimal-event.json")
	first, err := os.Open(fixturePath)
	require.NoError(t, err)
	firstResult, err := fixture.Import(ctx, pool, first)
	require.NoError(t, first.Close())
	require.NoError(t, err)
	require.True(t, firstResult.Changed)

	repository := event.NewRepository(pool)
	events, err := repository.ListPublic(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "magnitude-6-2-earthquake-near-sulawesi", events[0].Slug)
	require.Equal(t, "Magnitude 6.2 earthquake near Sulawesi", events[0].Title)
	require.Equal(t, -1.43, events[0].Latitude)
	require.Equal(t, 120.01, events[0].Longitude)
	require.Equal(t, 1, events[0].SourceCount)

	detail, err := repository.GetPublicBySlug(ctx, "magnitude-6-2-earthquake-near-sulawesi")
	require.NoError(t, err)
	require.Equal(t, 1, detail.Revision)
	require.NotEmpty(t, detail.Summary)

	second, err := os.Open(fixturePath)
	require.NoError(t, err)
	secondResult, err := fixture.Import(ctx, pool, second)
	require.NoError(t, second.Close())
	require.NoError(t, err)
	require.False(t, secondResult.Changed)

	var observationCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM observations").Scan(&observationCount))
	require.Equal(t, 1, observationCount)
	var revisionCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM event_revisions").Scan(&revisionCount))
	require.Equal(t, 1, revisionCount)
	var timelineCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM timeline_entries").Scan(&timelineCount))
	require.Equal(t, 1, timelineCount)
}
