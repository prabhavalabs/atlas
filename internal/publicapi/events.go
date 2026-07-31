// Package publicapi exposes anonymous, cacheable Atlas read endpoints.
package publicapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prabhavalabs/atlas/internal/event"
)

// EventReader is the narrow public event read boundary.
type EventReader interface {
	ListPublic(context.Context) ([]event.PublicEvent, error)
	GetPublicBySlug(context.Context, string) (event.PublicEvent, error)
}

// Register adds anonymous event endpoints to router.
func Register(router chi.Router, reader EventReader, now func() time.Time) {
	if now == nil {
		now = time.Now
	}
	router.Get("/api/v1/events", listEvents(reader, now))
	router.Get("/api/v1/events/{slug}", getEvent(reader, now))
}

type eventListResponse struct {
	Events      []event.PublicEvent `json:"events"`
	GeneratedAt time.Time           `json:"generatedAt"`
	Coverage    string              `json:"coverage"`
}

func listEvents(reader EventReader, now func() time.Time) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		events, err := reader.ListPublic(request.Context())
		if err != nil {
			writeProblem(writer, request, http.StatusServiceUnavailable, "events_unavailable", "Published events are temporarily unavailable")
			return
		}
		response := eventListResponse{Events: events, GeneratedAt: now().UTC(), Coverage: "fixture"}
		body, err := json.Marshal(response)
		if err != nil {
			writeProblem(writer, request, http.StatusInternalServerError, "response_encoding_failed", "The response could not be encoded")
			return
		}
		lastModified := response.GeneratedAt
		for _, publicEvent := range events {
			if publicEvent.VerifiedAt.After(lastModified) {
				lastModified = publicEvent.VerifiedAt
			}
		}
		writeCacheableJSON(writer, request, body, lastModified)
	}
}

func getEvent(reader EventReader, now func() time.Time) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		publicEvent, err := reader.GetPublicBySlug(request.Context(), chi.URLParam(request, "slug"))
		if errors.Is(err, event.ErrNotFound) {
			writeProblem(writer, request, http.StatusNotFound, "event_not_found", "Published event not found")
			return
		}
		if err != nil {
			writeProblem(writer, request, http.StatusServiceUnavailable, "event_unavailable", "The event is temporarily unavailable")
			return
		}
		body, err := json.Marshal(struct {
			Event       event.PublicEvent `json:"event"`
			GeneratedAt time.Time         `json:"generatedAt"`
		}{Event: publicEvent, GeneratedAt: now().UTC()})
		if err != nil {
			writeProblem(writer, request, http.StatusInternalServerError, "response_encoding_failed", "The response could not be encoded")
			return
		}
		writeCacheableJSON(writer, request, body, publicEvent.VerifiedAt)
	}
}

func writeCacheableJSON(writer http.ResponseWriter, request *http.Request, body []byte, lastModified time.Time) {
	digest := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(digest[:]) + `"`
	writer.Header().Set("ETag", etag)
	writer.Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	writer.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	if request.Header.Get("If-None-Match") == etag {
		writer.WriteHeader(http.StatusNotModified)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(body)
}

func writeProblem(writer http.ResponseWriter, request *http.Request, status int, code string, title string) {
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"type":     "https://atlas.prabhavalabs.com/problems/" + code,
		"title":    title,
		"status":   status,
		"code":     code,
		"instance": request.URL.Path,
	})
}
