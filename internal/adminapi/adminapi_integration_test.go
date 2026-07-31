//go:build integration

package adminapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prabhavalabs/atlas/internal/adminapi"
	"github.com/prabhavalabs/atlas/internal/event"
	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/prabhavalabs/atlas/internal/source/fixture"
	"github.com/stretchr/testify/require"
)

func TestAdminLoginUsesUniformFailureAndSecureSession(t *testing.T) {
	router, _, _ := setupAdminAPI(t, identity.RoleAdministrator)

	unknown := performJSON(t, router, http.MethodPost, "/api/v1/admin/session", map[string]string{
		"email": "unknown@example.org", "password": "wrong password value",
	}, nil)
	wrong := performJSON(t, router, http.MethodPost, "/api/v1/admin/session", map[string]string{
		"email": "admin@example.org", "password": "wrong password value",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, unknown.Code)
	require.Equal(t, unknown.Body.String(), wrong.Body.String())
	require.Contains(t, unknown.Body.String(), "invalid_credentials")

	success := performJSON(t, router, http.MethodPost, "/api/v1/admin/session", map[string]string{
		"email": "admin@example.org", "password": "a strong and memorable passphrase",
	}, nil)
	require.Equal(t, http.StatusOK, success.Code)
	require.Contains(t, success.Body.String(), `"role":"administrator"`)
	require.Contains(t, success.Body.String(), "csrfToken")
	cookies := success.Result().Cookies()
	require.Len(t, cookies, 2)
	sessionCookie := cookieByName(t, cookies, identity.SessionCookieName())
	csrfCookie := cookieByName(t, cookies, identity.CSRFCookieName())
	require.True(t, sessionCookie.HttpOnly)
	require.True(t, sessionCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)
	require.False(t, csrfCookie.HttpOnly)
	require.True(t, csrfCookie.Secure)
	require.Equal(t, http.SameSiteStrictMode, csrfCookie.SameSite)
	require.Equal(t, "atlas.example.org", csrfCookie.Domain)
}

func TestAdminMutationRequiresSessionCSRFRoleAndCreatesAudit(t *testing.T) {
	router, repository, pool := setupAdminAPI(t, identity.RoleAdministrator)
	publicEvents, err := repository.ListPublic(context.Background())
	require.NoError(t, err)
	require.Len(t, publicEvents, 1)
	eventID := publicEvents[0].ID

	unauthorized := performJSON(t, router, http.MethodPatch, "/api/v1/admin/events/"+eventID+"/title", map[string]any{
		"title": "Updated earthquake title", "reason": "Verified against the latest official bulletin.", "expectedRevision": 1,
	}, nil)
	require.Equal(t, http.StatusUnauthorized, unauthorized.Code)

	sessionCookie, csrfToken := login(t, router)
	missingCSRF := performJSON(t, router, http.MethodPatch, "/api/v1/admin/events/"+eventID+"/title", map[string]any{
		"title": "Updated earthquake title", "reason": "Verified against the latest official bulletin.", "expectedRevision": 1,
	}, map[string]string{"Cookie": sessionCookie.String()})
	require.Equal(t, http.StatusForbidden, missingCSRF.Code)

	updated := performJSON(t, router, http.MethodPatch, "/api/v1/admin/events/"+eventID+"/title", map[string]any{
		"title": "Updated earthquake title", "reason": "Verified against the latest official bulletin.", "expectedRevision": 1,
	}, map[string]string{"Cookie": sessionCookie.String(), "X-CSRF-Token": csrfToken, "X-Request-ID": "test-request-id"})
	require.Equal(t, http.StatusOK, updated.Code)
	require.Contains(t, updated.Body.String(), `"revision":2`)

	publicEvent, err := repository.GetPublicBySlug(context.Background(), "magnitude-6-2-earthquake-near-sulawesi")
	require.NoError(t, err)
	require.Equal(t, "Updated earthquake title", publicEvent.Title)
	require.Equal(t, 2, publicEvent.Revision)

	var auditCount int
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM audit_log
		WHERE action = 'event.title.updated' AND target_id = $1 AND request_id = 'test-request-id'
	`, eventID).Scan(&auditCount))
	require.Equal(t, 1, auditCount)
}

func TestViewerCannotEditEvent(t *testing.T) {
	router, repository, _ := setupAdminAPI(t, identity.RoleViewer)
	publicEvents, err := repository.ListPublic(context.Background())
	require.NoError(t, err)
	eventID := publicEvents[0].ID
	cookie, csrfToken := login(t, router)

	response := performJSON(t, router, http.MethodPatch, "/api/v1/admin/events/"+eventID+"/title", map[string]any{
		"title": "Viewer must not publish", "reason": "This request must be rejected by role policy.", "expectedRevision": 1,
	}, map[string]string{"Cookie": cookie.String(), "X-CSRF-Token": csrfToken})

	require.Equal(t, http.StatusForbidden, response.Code)
	require.Contains(t, response.Body.String(), "insufficient_role")
}

func setupAdminAPI(t *testing.T, role identity.Role) (http.Handler, *event.Repository, *pgxpool.Pool) {
	t.Helper()
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))
	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE audit_log, sessions, users, timeline_entries, event_observations, event_revisions, events, observations, sources RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	fixtureFile, err := os.Open(filepath.Join("..", "..", "testdata", "sources", "fixture", "minimal-event.json"))
	require.NoError(t, err)
	_, err = fixture.Import(ctx, pool, fixtureFile)
	require.NoError(t, fixtureFile.Close())
	require.NoError(t, err)

	identityRepository := identity.NewRepository(pool)
	passwordHash, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)
	_, err = identityRepository.CreateUser(ctx, "admin@example.org", "Atlas Admin", passwordHash, role)
	require.NoError(t, err)

	eventRepository := event.NewRepository(pool)
	router := chi.NewRouter()
	adminapi.Register(router, identityRepository, eventRepository, adminapi.Options{
		SecureCookie:     true,
		CSRFCookieDomain: "atlas.example.org",
		SessionTTL:       time.Hour,
		Now: func() time.Time {
			return time.Date(2026, 7, 31, 20, 0, 0, 0, time.UTC)
		},
	})
	return router, eventRepository, pool
}

func login(t *testing.T, router http.Handler) (*http.Cookie, string) {
	t.Helper()
	response := performJSON(t, router, http.MethodPost, "/api/v1/admin/session", map[string]string{
		"email": "admin@example.org", "password": "a strong and memorable passphrase",
	}, nil)
	require.Equal(t, http.StatusOK, response.Code)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 2)
	sessionCookie := cookieByName(t, cookies, identity.SessionCookieName())
	csrfCookie := cookieByName(t, cookies, identity.CSRFCookieName())
	return sessionCookie, csrfCookie.Value
}

func cookieByName(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

func performJSON(t *testing.T, router http.Handler, method string, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
