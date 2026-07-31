//go:build integration

package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/prabhavalabs/atlas/internal/platform/app"
	"github.com/prabhavalabs/atlas/internal/platform/config"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/prabhavalabs/atlas/internal/source/fixture"
	"github.com/stretchr/testify/require"
)

func TestHandlerAssemblesPublicAdminAndWebRoutes(t *testing.T) {
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))
	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE audit_log, sessions, users, timeline_entries, event_observations, event_revisions, events, observations, sources RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	fixtureFile, err := os.Open(filepath.Join("..", "..", "..", "testdata", "sources", "fixture", "minimal-event.json"))
	require.NoError(t, err)
	_, err = fixture.Import(ctx, pool, fixtureFile)
	require.NoError(t, fixtureFile.Close())
	require.NoError(t, err)

	handler := app.NewHandler(app.HandlerOptions{
		Pool: pool,
		Config: config.Config{
			AllowedOrigins:      []string{"https://atlas.example.org"},
			SessionCookieSecure: true,
			CSRFCookieDomain:    "atlas.example.org",
			SessionTTL:          time.Hour,
		},
		Web: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<title>Atlas web</title>")}},
		Now: func() time.Time {
			return time.Date(2026, time.July, 31, 20, 5, 0, 0, time.UTC)
		},
	})

	publicRecorder := httptest.NewRecorder()
	handler.ServeHTTP(publicRecorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events", nil))
	require.Equal(t, http.StatusOK, publicRecorder.Code)
	require.Contains(t, publicRecorder.Body.String(), "Magnitude 6.2 earthquake")

	adminRecorder := httptest.NewRecorder()
	handler.ServeHTTP(adminRecorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/admin/session", nil))
	require.Equal(t, http.StatusUnauthorized, adminRecorder.Code)

	webRecorder := httptest.NewRecorder()
	handler.ServeHTTP(webRecorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil))
	require.Equal(t, http.StatusOK, webRecorder.Code)
	require.Contains(t, webRecorder.Body.String(), "Atlas web")
}
