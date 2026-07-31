//go:build integration

package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/stretchr/testify/require"
)

func TestMigrateCreatesRequiredExtensionsAndMetadata(t *testing.T) {
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL, "ATLAS_TEST_DATABASE_URL must reference a disposable database")

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	var postGISVersion string
	require.NoError(t, pool.QueryRow(ctx, "SELECT postgis_full_version()").Scan(&postGISVersion))
	require.Contains(t, postGISVersion, "POSTGIS")

	var extensionCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM pg_extension WHERE extname IN ('postgis', 'pg_trgm', 'citext')").Scan(&extensionCount))
	require.Equal(t, 3, extensionCount)

	var applicationVersion string
	require.NoError(t, pool.QueryRow(ctx, "SELECT application_version FROM schema_metadata WHERE singleton = TRUE").Scan(&applicationVersion))
	require.Equal(t, "foundation", applicationVersion)
}
