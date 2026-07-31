package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

const (
	sessionCookieName = "atlas_session"
	csrfCookieName    = "atlas_admin_csrf"
)

// SessionTokens contains browser-facing secrets and one-way database hashes.
type SessionTokens struct {
	SessionToken string
	SessionHash  []byte
	CSRFToken    string
	CSRFHash     []byte
}

// NewSessionTokens generates independent high-entropy session and CSRF secrets.
func NewSessionTokens() (SessionTokens, error) {
	sessionToken, err := randomToken()
	if err != nil {
		return SessionTokens{}, fmt.Errorf("generate session token: %w", err)
	}
	csrfToken, err := randomToken()
	if err != nil {
		return SessionTokens{}, fmt.Errorf("generate CSRF token: %w", err)
	}
	return SessionTokens{
		SessionToken: sessionToken,
		SessionHash:  secretHash(sessionToken),
		CSRFToken:    csrfToken,
		CSRFHash:     secretHash(csrfToken),
	}, nil
}

// MatchesSecret compares a raw token with its SHA-256 database representation.
func MatchesSecret(raw string, expectedHash []byte) bool {
	actualHash := secretHash(raw)
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

// SessionCookie constructs the only browser cookie used by administrator sessions.
func SessionCookie(token string, expires time.Time, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/api/v1/admin",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// CSRFCookie persists the double-submit token where the frontend can read it
// after a reload. The API still validates the value against the session's hash.
func CSRFCookie(token string, expires time.Time, secure bool, domain string) *http.Cookie {
	return &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		Domain:   domain,
		Expires:  expires,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
}

// ExpiredSessionCookie instructs the browser to remove its admin session.
func ExpiredSessionCookie(secure bool) *http.Cookie {
	cookie := SessionCookie("", time.Unix(1, 0).UTC(), secure)
	cookie.MaxAge = -1
	return cookie
}

// ExpiredCSRFCookie instructs the browser to remove its readable CSRF token.
func ExpiredCSRFCookie(secure bool, domain string) *http.Cookie {
	cookie := CSRFCookie("", time.Unix(1, 0).UTC(), secure, domain)
	cookie.MaxAge = -1
	return cookie
}

// SessionCookieName is the stable cookie name used by middleware.
func SessionCookieName() string {
	return sessionCookieName
}

// CSRFCookieName is the stable readable double-submit cookie name.
func CSRFCookieName() string {
	return csrfCookieName
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func secretHash(raw string) []byte {
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}
