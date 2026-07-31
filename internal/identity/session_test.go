package identity_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestSessionTokensAreOpaqueAndMatchOnlyTheirHashes(t *testing.T) {
	tokens, err := identity.NewSessionTokens()
	require.NoError(t, err)

	require.NotEmpty(t, tokens.SessionToken)
	require.NotEmpty(t, tokens.CSRFToken)
	require.NotEqual(t, tokens.SessionToken, tokens.CSRFToken)
	require.True(t, identity.MatchesSecret(tokens.SessionToken, tokens.SessionHash))
	require.True(t, identity.MatchesSecret(tokens.CSRFToken, tokens.CSRFHash))
	require.False(t, identity.MatchesSecret("wrong-token", tokens.SessionHash))
}

func TestSessionCookieUsesSecureAdministrativeDefaults(t *testing.T) {
	expires := time.Date(2026, time.August, 1, 8, 0, 0, 0, time.UTC)
	cookie := identity.SessionCookie("opaque-token", expires, true)

	require.Equal(t, "atlas_session", cookie.Name)
	require.Equal(t, "opaque-token", cookie.Value)
	require.Equal(t, "/api/v1/admin", cookie.Path)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	require.Equal(t, expires, cookie.Expires)
}

func TestExpiredSessionCookieClearsTheBrowserValue(t *testing.T) {
	cookie := identity.ExpiredSessionCookie(true)

	require.Empty(t, cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
	require.True(t, cookie.Expires.Before(time.Now()))
}

func TestCSRFCookieIsReadableAcrossAtlasFrontendAndAPIHosts(t *testing.T) {
	expires := time.Date(2026, time.August, 1, 8, 0, 0, 0, time.UTC)
	cookie := identity.CSRFCookie("csrf-token", expires, true, "atlas.example.org")

	require.Equal(t, "atlas_admin_csrf", cookie.Name)
	require.Equal(t, "csrf-token", cookie.Value)
	require.Equal(t, "/", cookie.Path)
	require.Equal(t, "atlas.example.org", cookie.Domain)
	require.False(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	require.Equal(t, expires, cookie.Expires)
}

func TestExpiredCSRFCookieClearsTheBrowserValue(t *testing.T) {
	cookie := identity.ExpiredCSRFCookie(true, "atlas.example.org")

	require.Empty(t, cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
	require.Equal(t, "atlas.example.org", cookie.Domain)
	require.True(t, cookie.Expires.Before(time.Now()))
}
