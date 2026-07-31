package config_test

import (
	"testing"

	"github.com/prabhavalabs/atlas/internal/platform/config"
	"github.com/stretchr/testify/require"
)

func validEnvironment() map[string]string {
	return map[string]string{
		"ATLAS_DATABASE_URL": "postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable",
		"ATLAS_PUBLIC_URL":   "http://localhost:5173",
		"ATLAS_API_URL":      "http://localhost:8081",
	}
}

func lookup(environment map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		value, ok := environment[key]
		return value, ok
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	environment := validEnvironment()
	delete(environment, "ATLAS_DATABASE_URL")

	_, err := config.Load(lookup(environment))

	require.ErrorContains(t, err, "ATLAS_DATABASE_URL")
}

func TestLoadRejectsInvalidPublicURL(t *testing.T) {
	environment := validEnvironment()
	environment["ATLAS_PUBLIC_URL"] = "://not-a-url"

	_, err := config.Load(lookup(environment))

	require.ErrorContains(t, err, "ATLAS_PUBLIC_URL")
}

func TestLoadRejectsInvalidLogLevel(t *testing.T) {
	environment := validEnvironment()
	environment["ATLAS_LOG_LEVEL"] = "verbose"

	_, err := config.Load(lookup(environment))

	require.ErrorContains(t, err, "ATLAS_LOG_LEVEL")
}

func TestLoadRejectsUnsupportedCookieSameSiteMode(t *testing.T) {
	environment := validEnvironment()
	environment["ATLAS_COOKIE_SAME_SITE"] = "None"

	_, err := config.Load(lookup(environment))

	require.ErrorContains(t, err, "ATLAS_COOKIE_SAME_SITE")
}

func TestLoadAcceptsMinimalConfigurationAndAppliesDefaults(t *testing.T) {
	environment := validEnvironment()

	configuration, err := config.Load(lookup(environment))

	require.NoError(t, err)
	require.Equal(t, "development", configuration.Environment)
	require.Equal(t, "127.0.0.1:8081", configuration.HTTPAddress)
	require.Equal(t, "web/dist", configuration.WebRoot)
	require.Equal(t, "info", configuration.LogLevel)
	require.Equal(t, "Lax", configuration.CookieSameSite)
	require.Equal(t, "https://openrouter.ai/api/v1", configuration.OpenRouterBaseURL)
	require.Empty(t, configuration.OpenRouterAPIKey)
	require.Empty(t, configuration.CSRFCookieDomain)
}

func TestLoadDefaultsCSRFCookieDomainToPublicProductionHost(t *testing.T) {
	environment := validEnvironment()
	environment["ATLAS_ENVIRONMENT"] = "production"
	environment["ATLAS_COOKIE_SECURE"] = "true"
	environment["ATLAS_PUBLIC_URL"] = "https://atlas.example.org"
	environment["ATLAS_API_URL"] = "https://api.atlas.example.org"

	configuration, err := config.Load(lookup(environment))

	require.NoError(t, err)
	require.Equal(t, "atlas.example.org", configuration.CSRFCookieDomain)
}

func TestLoadRejectsCSRFCookieDomainOutsidePublicAndAPIHosts(t *testing.T) {
	environment := validEnvironment()
	environment["ATLAS_PUBLIC_URL"] = "https://atlas.example.org"
	environment["ATLAS_API_URL"] = "https://api.atlas.example.org"
	environment["ATLAS_CSRF_COOKIE_DOMAIN"] = "unrelated.example.net"

	_, err := config.Load(lookup(environment))

	require.ErrorContains(t, err, "ATLAS_CSRF_COOKIE_DOMAIN")
}
