// Package adminapi exposes authenticated, audited editorial endpoints.
package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prabhavalabs/atlas/internal/event"
	"github.com/prabhavalabs/atlas/internal/identity"
)

const maximumAdminBodyBytes = 64 << 10

// Options controls secure administrator session behavior.
type Options struct {
	SecureCookie     bool
	CSRFCookieDomain string
	SessionTTL       time.Duration
	Now              func() time.Time
}

// Register adds all administrator endpoints to router.
func Register(router chi.Router, identities *identity.Repository, events *event.Repository, options Options) {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.SessionTTL <= 0 {
		options.SessionTTL = 12 * time.Hour
	}
	dummyHash, err := identity.HashPassword("constant invalid login password")
	if err != nil {
		panic("create login comparison hash")
	}

	handler := handler{
		identities:    identities,
		events:        events,
		options:       options,
		dummyHash:     dummyHash,
		loginAttempts: newAttemptLimiter(5, 15*time.Minute),
	}
	router.Post("/api/v1/admin/session", handler.login)
	router.With(handler.requireSession(false)).Get("/api/v1/admin/session", handler.currentSession)
	router.With(handler.requireSession(true)).Delete("/api/v1/admin/session", handler.logout)
	router.With(handler.requireSession(true), requireRole(identity.RoleAdministrator, identity.RoleEditor)).Patch(
		"/api/v1/admin/events/{id}/title",
		handler.updateEventTitle,
	)
}

type handler struct {
	identities    *identity.Repository
	events        *event.Repository
	options       Options
	dummyHash     string
	loginAttempts *attemptLimiter
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (handler handler) login(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "no-store")
	var input loginInput
	if err := decodeJSON(writer, request, &input); err != nil {
		writeProblem(writer, request, http.StatusBadRequest, "invalid_request", "The login request is invalid")
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(input.Email); err != nil || len(input.Password) > 1024 {
		handler.invalidLogin(writer, request)
		return
	}
	attemptKey := clientIP(request) + "|" + input.Email
	if !handler.loginAttempts.Allow(attemptKey, handler.options.Now()) {
		writeProblem(writer, request, http.StatusTooManyRequests, "login_rate_limited", "Too many login attempts; try again later")
		return
	}

	credentialUser, err := handler.identities.FindCredentialUserByEmail(request.Context(), input.Email)
	if errors.Is(err, identity.ErrInvalidCredentials) {
		_, _ = identity.VerifyPassword(input.Password, handler.dummyHash)
		handler.invalidLogin(writer, request)
		return
	}
	if err != nil {
		writeProblem(writer, request, http.StatusServiceUnavailable, "identity_unavailable", "Administrator identity is temporarily unavailable")
		return
	}
	verified, err := identity.VerifyPassword(input.Password, credentialUser.PasswordHash)
	if err != nil || !verified || !credentialUser.Active {
		handler.invalidLogin(writer, request)
		return
	}

	if existingCookie, cookieErr := request.Cookie(identity.SessionCookieName()); cookieErr == nil {
		if existing, authenticationErr := handler.identities.AuthenticateSession(request.Context(), existingCookie.Value, handler.options.Now()); authenticationErr == nil {
			_ = handler.identities.RevokeSession(request.Context(), existing.SessionID, handler.options.Now())
		}
	}

	tokens, err := identity.NewSessionTokens()
	if err != nil {
		writeProblem(writer, request, http.StatusInternalServerError, "session_creation_failed", "A secure session could not be created")
		return
	}
	expiresAt := handler.options.Now().Add(handler.options.SessionTTL)
	_, err = handler.identities.CreateSession(request.Context(), credentialUser.ID, tokens, expiresAt, handler.options.Now())
	if err != nil {
		writeProblem(writer, request, http.StatusServiceUnavailable, "session_unavailable", "A secure session could not be stored")
		return
	}
	handler.loginAttempts.Reset(attemptKey)
	http.SetCookie(writer, identity.SessionCookie(tokens.SessionToken, expiresAt, handler.options.SecureCookie))
	http.SetCookie(writer, identity.CSRFCookie(tokens.CSRFToken, expiresAt, handler.options.SecureCookie, handler.options.CSRFCookieDomain))
	writeJSON(writer, http.StatusOK, map[string]any{
		"user":      credentialUser.User,
		"csrfToken": tokens.CSRFToken,
		"expiresAt": expiresAt,
	})
}

func (handler handler) invalidLogin(writer http.ResponseWriter, request *http.Request) {
	writeProblem(writer, request, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
}

type sessionContextKey struct{}

func (handler handler) requireSession(requireCSRF bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Cache-Control", "no-store")
			cookie, err := request.Cookie(identity.SessionCookieName())
			if err != nil {
				writeProblem(writer, request, http.StatusUnauthorized, "authentication_required", "Administrator authentication is required")
				return
			}
			session, err := handler.identities.AuthenticateSession(request.Context(), cookie.Value, handler.options.Now())
			if err != nil {
				http.SetCookie(writer, identity.ExpiredSessionCookie(handler.options.SecureCookie))
				writeProblem(writer, request, http.StatusUnauthorized, "authentication_required", "Administrator authentication is required")
				return
			}
			if requireCSRF && !identity.MatchesSecret(request.Header.Get("X-CSRF-Token"), session.CSRFHash) {
				writeProblem(writer, request, http.StatusForbidden, "csrf_validation_failed", "The request could not be verified")
				return
			}
			ctx := context.WithValue(request.Context(), sessionContextKey{}, session)
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

func requireRole(roles ...identity.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			session, ok := request.Context().Value(sessionContextKey{}).(identity.AuthenticatedSession)
			if !ok {
				writeProblem(writer, request, http.StatusUnauthorized, "authentication_required", "Administrator authentication is required")
				return
			}
			for _, role := range roles {
				if session.User.Role == role {
					next.ServeHTTP(writer, request)
					return
				}
			}
			writeProblem(writer, request, http.StatusForbidden, "insufficient_role", "Your role cannot perform this action")
		})
	}
}

func (handler handler) currentSession(writer http.ResponseWriter, request *http.Request) {
	session := request.Context().Value(sessionContextKey{}).(identity.AuthenticatedSession)
	writeJSON(writer, http.StatusOK, map[string]any{"user": session.User, "expiresAt": session.ExpiresAt})
}

func (handler handler) logout(writer http.ResponseWriter, request *http.Request) {
	session := request.Context().Value(sessionContextKey{}).(identity.AuthenticatedSession)
	if err := handler.identities.RevokeSession(request.Context(), session.SessionID, handler.options.Now()); err != nil {
		writeProblem(writer, request, http.StatusServiceUnavailable, "logout_failed", "The session could not be revoked")
		return
	}
	http.SetCookie(writer, identity.ExpiredSessionCookie(handler.options.SecureCookie))
	http.SetCookie(writer, identity.ExpiredCSRFCookie(handler.options.SecureCookie, handler.options.CSRFCookieDomain))
	writer.WriteHeader(http.StatusNoContent)
}

type updateTitleInput struct {
	Title            string `json:"title"`
	Reason           string `json:"reason"`
	ExpectedRevision int    `json:"expectedRevision"`
}

func (handler handler) updateEventTitle(writer http.ResponseWriter, request *http.Request) {
	var input updateTitleInput
	if err := decodeJSON(writer, request, &input); err != nil {
		writeProblem(writer, request, http.StatusBadRequest, "invalid_request", "The event update request is invalid")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Title) < 5 || len(input.Title) > 200 || len(input.Reason) < 10 || len(input.Reason) > 500 || input.ExpectedRevision < 1 {
		writeProblem(writer, request, http.StatusUnprocessableEntity, "validation_failed", "Title, reason, or revision is invalid")
		return
	}
	eventID, err := uuid.Parse(chi.URLParam(request, "id"))
	if err != nil {
		writeProblem(writer, request, http.StatusNotFound, "event_not_found", "Event not found")
		return
	}
	session := request.Context().Value(sessionContextKey{}).(identity.AuthenticatedSession)
	updated, err := handler.events.UpdatePublishedTitle(request.Context(), event.UpdateTitleInput{
		EventID:          eventID,
		Title:            input.Title,
		Reason:           input.Reason,
		ExpectedRevision: input.ExpectedRevision,
		ActorID:          session.User.ID,
		RequestID:        request.Header.Get("X-Request-ID"),
		Now:              handler.options.Now(),
	})
	if errors.Is(err, event.ErrRevisionConflict) {
		writeProblem(writer, request, http.StatusConflict, "revision_conflict", "The event changed; reload before editing")
		return
	}
	if errors.Is(err, event.ErrNotFound) {
		writeProblem(writer, request, http.StatusNotFound, "event_not_found", "Event not found")
		return
	}
	if err != nil {
		writeProblem(writer, request, http.StatusServiceUnavailable, "event_update_failed", "The event could not be updated")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"event": updated})
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, target any) error {
	request.Body = http.MaxBytesReader(writer, request.Body, maximumAdminBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func clientIP(request *http.Request) string {
	if value := strings.TrimSpace(request.Header.Get("CF-Connecting-IP")); value != "" {
		return value
	}
	return request.RemoteAddr
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

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
