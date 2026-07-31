package publicapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prabhavalabs/atlas/internal/event"
	"github.com/prabhavalabs/atlas/internal/publicapi"
	"github.com/stretchr/testify/require"
)

type eventReader struct {
	events []event.PublicEvent
}

func (reader eventReader) ListPublic(context.Context) ([]event.PublicEvent, error) {
	return reader.events, nil
}

func (reader eventReader) GetPublicBySlug(_ context.Context, slug string) (event.PublicEvent, error) {
	for _, publicEvent := range reader.events {
		if publicEvent.Slug == slug {
			return publicEvent, nil
		}
	}
	return event.PublicEvent{}, event.ErrNotFound
}

func fixtureEvent() event.PublicEvent {
	return event.PublicEvent{
		ID:           "61d8ed9c-1ce8-47c3-97f1-88bead4e15bc",
		Slug:         "magnitude-6-2-earthquake-near-sulawesi",
		Type:         "earthquake",
		Lifecycle:    "active",
		Priority:     "high",
		Confidence:   "high",
		Title:        "Magnitude 6.2 earthquake near Sulawesi",
		Summary:      "Initial assessments remain in progress.",
		LocationName: "Sulawesi, Indonesia",
		Latitude:     -1.43,
		Longitude:    120.01,
		StartedAt:    time.Date(2026, 7, 31, 19, 41, 0, 0, time.UTC),
		VerifiedAt:   time.Date(2026, 7, 31, 20, 4, 0, 0, time.UTC),
		SourceCount:  1,
		Revision:     1,
	}
}

func newPublicRouter(t *testing.T, events []event.PublicEvent) http.Handler {
	t.Helper()
	router := chi.NewRouter()
	publicapi.Register(router, eventReader{events: events}, func() time.Time {
		return time.Date(2026, 7, 31, 20, 30, 0, 0, time.UTC)
	})
	return router
}

func TestListEventsIsAnonymousAndReturnsCoverageMetadata(t *testing.T) {
	router := newPublicRouter(t, []event.PublicEvent{fixtureEvent()})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Magnitude 6.2 earthquake near Sulawesi")
	require.Contains(t, recorder.Body.String(), `"coverage":"fixture"`)
	require.NotEmpty(t, recorder.Header().Get("ETag"))
	require.NotEmpty(t, recorder.Header().Get("Last-Modified"))
}

func TestListEventsHonorsETagRevalidation(t *testing.T) {
	router := newPublicRouter(t, []event.PublicEvent{fixtureEvent()})
	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events", nil))

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events", nil)
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	second := httptest.NewRecorder()
	router.ServeHTTP(second, request)

	require.Equal(t, http.StatusNotModified, second.Code)
	require.Empty(t, second.Body.String())
}

func TestGetEventReturnsPublishedNarrative(t *testing.T) {
	router := newPublicRouter(t, []event.PublicEvent{fixtureEvent()})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events/magnitude-6-2-earthquake-near-sulawesi", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Initial assessments remain in progress")
	require.Contains(t, recorder.Body.String(), `"revision":1`)
}

func TestGetEventReturnsProblemForUnknownSlug(t *testing.T) {
	router := newPublicRouter(t, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/events/unknown", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), "event_not_found")
}
