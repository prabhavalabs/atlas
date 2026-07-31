package httpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prabhavalabs/atlas/internal/platform/buildinfo"
	"github.com/prabhavalabs/atlas/internal/platform/httpserver"
	"github.com/stretchr/testify/require"
)

func newServer(t *testing.T, ready func(context.Context) error) http.Handler {
	t.Helper()
	return httpserver.New(httpserver.Options{
		Build: buildinfo.Info{
			Version:   "v0.1.0-test",
			Commit:    "abc123",
			BuildTime: "2026-07-31T20:00:00Z",
		},
		Ready:          ready,
		AllowedOrigins: []string{"https://atlas.prabhavalabs.com"},
		Now: func() time.Time {
			return time.Date(2026, time.July, 31, 20, 30, 0, 0, time.UTC)
		},
	})
}

func TestLivenessDoesNotDependOnDatabase(t *testing.T) {
	readyCalls := 0
	server := newServer(t, func(context.Context) error {
		readyCalls++
		return errors.New("database unavailable")
	})
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/live", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Zero(t, readyCalls)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestReadinessReturnsProblemWhenDependencyIsUnavailable(t *testing.T) {
	server := newServer(t, func(context.Context) error { return errors.New("secret database details") })
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
	require.NotContains(t, recorder.Body.String(), "secret database details")
	require.Contains(t, recorder.Body.String(), "dependency_unavailable")
}

func TestMetaExposesBuildAndGenerationTime(t *testing.T) {
	server := newServer(t, func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/meta", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "v0.1.0-test", body["version"])
	require.Equal(t, "abc123", body["commit"])
	require.Equal(t, "2026-07-31T20:30:00Z", body["generatedAt"])
}

func TestUnknownRouteReturnsStableProblem(t *testing.T) {
	server := newServer(t, func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/does-not-exist", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), "route_not_found")
}

func TestMethodRejectionReturnsStableProblem(t *testing.T) {
	server := newServer(t, func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/health/live", nil))

	require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	require.Contains(t, recorder.Body.String(), "method_not_allowed")
}

func TestRequestIDIsGeneratedAndSecurityHeadersArePresent(t *testing.T) {
	server := newServer(t, func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/live", nil))

	require.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
	require.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
}

func TestContentSecurityPolicyAllowsConfiguredAPIAndMapTiles(t *testing.T) {
	server := httpserver.New(httpserver.Options{APIOrigin: "https://api.atlas.example.org"})
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/live", nil))

	policy := recorder.Header().Get("Content-Security-Policy")
	require.Contains(t, policy, "connect-src 'self' https://api.atlas.example.org https://tiles.openfreemap.org")
	require.Contains(t, policy, "worker-src 'self' blob:")
}

func TestMetricsRequireOperatorToken(t *testing.T) {
	server := httpserver.New(httpserver.Options{MetricsToken: "metrics-secret"})

	unauthorized := httptest.NewRecorder()
	server.ServeHTTP(unauthorized, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusUnauthorized, unauthorized.Code)
	require.NotContains(t, unauthorized.Body.String(), "metrics-secret")

	authorizedRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer metrics-secret")
	authorized := httptest.NewRecorder()
	server.ServeHTTP(authorized, authorizedRequest)
	require.Equal(t, http.StatusOK, authorized.Code)
	require.Contains(t, authorized.Header().Get("Content-Type"), "text/plain")
	require.Contains(t, authorized.Body.String(), "atlas_up 1")
}

func TestRequestIDIsAvailableToRegisteredHandlers(t *testing.T) {
	var handlerRequestID string
	server := httpserver.New(httpserver.Options{
		Register: func(router chi.Router) {
			router.Get("/capture-request-id", func(writer http.ResponseWriter, request *http.Request) {
				handlerRequestID = request.Header.Get("X-Request-ID")
				writer.WriteHeader(http.StatusNoContent)
			})
		},
	})
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/capture-request-id", nil))

	require.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
	require.Equal(t, recorder.Header().Get("X-Request-ID"), handlerRequestID)
}

func TestAllowedCORSOriginReceivesCredentialHeaders(t *testing.T) {
	server := newServer(t, func(context.Context) error { return nil })
	request := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/api/v1/meta", nil)
	request.Header.Set("Origin", "https://atlas.prabhavalabs.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, "https://atlas.prabhavalabs.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
}

func TestWebApplicationServesAssetsAndSPAFallbackWithoutMaskingAPIProblems(t *testing.T) {
	web := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<!doctype html><title>Atlas application</title>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('atlas')")},
	}
	server := httpserver.New(httpserver.Options{Web: web})

	root := httptest.NewRecorder()
	server.ServeHTTP(root, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, root.Code)
	require.Contains(t, root.Body.String(), "Atlas application")
	require.Contains(t, root.Header().Get("Content-Type"), "text/html")
	require.Equal(t, "no-cache", root.Header().Get("Cache-Control"))

	asset := httptest.NewRecorder()
	server.ServeHTTP(asset, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/assets/app.js", nil))
	require.Equal(t, http.StatusOK, asset.Code)
	require.Equal(t, "public, max-age=31536000, immutable", asset.Header().Get("Cache-Control"))

	admin := httptest.NewRecorder()
	server.ServeHTTP(admin, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil))
	require.Equal(t, http.StatusOK, admin.Code)
	require.Contains(t, admin.Body.String(), "Atlas application")

	unknownAPI := httptest.NewRecorder()
	server.ServeHTTP(unknownAPI, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/unknown", nil))
	require.Equal(t, http.StatusNotFound, unknownAPI.Code)
	require.Contains(t, unknownAPI.Body.String(), "route_not_found")
}
