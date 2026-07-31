// Package config owns all Atlas runtime configuration and validation.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LookupFunc mirrors os.LookupEnv and keeps configuration loading testable.
type LookupFunc func(string) (string, bool)

// Config is the validated runtime configuration for the Atlas process.
type Config struct {
	Environment          string
	HTTPAddress          string
	WebRoot              string
	DatabaseURL          string
	PublicURL            string
	APIURL               string
	AllowedOrigins       []string
	LogLevel             string
	SessionCookieSecure  bool
	CSRFCookieDomain     string
	CookieSameSite       string
	SessionTTL           time.Duration
	MetricsToken         string
	OpenRouterAPIKey     string
	OpenRouterModel      string
	OpenRouterBaseURL    string
	OpenRouterMaxCostUSD float64
}

// Load reads all process settings through lookup and returns a validated Config.
func Load(lookup LookupFunc) (Config, error) {
	configuration := Config{
		Environment:          valueOrDefault(lookup, "ATLAS_ENVIRONMENT", "development"),
		HTTPAddress:          valueOrDefault(lookup, "ATLAS_HTTP_ADDRESS", "127.0.0.1:8081"),
		WebRoot:              valueOrDefault(lookup, "ATLAS_WEB_ROOT", "web/dist"),
		DatabaseURL:          value(lookup, "ATLAS_DATABASE_URL"),
		PublicURL:            value(lookup, "ATLAS_PUBLIC_URL"),
		APIURL:               value(lookup, "ATLAS_API_URL"),
		LogLevel:             valueOrDefault(lookup, "ATLAS_LOG_LEVEL", "info"),
		CSRFCookieDomain:     strings.TrimPrefix(value(lookup, "ATLAS_CSRF_COOKIE_DOMAIN"), "."),
		CookieSameSite:       valueOrDefault(lookup, "ATLAS_COOKIE_SAME_SITE", "Lax"),
		MetricsToken:         value(lookup, "ATLAS_METRICS_TOKEN"),
		OpenRouterAPIKey:     value(lookup, "ATLAS_OPENROUTER_API_KEY"),
		OpenRouterModel:      value(lookup, "ATLAS_OPENROUTER_MODEL"),
		OpenRouterBaseURL:    valueOrDefault(lookup, "ATLAS_OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		SessionTTL:           12 * time.Hour,
		OpenRouterMaxCostUSD: 0,
	}

	secureCookie, err := boolValue(lookup, "ATLAS_COOKIE_SECURE", configuration.Environment == "production")
	if err != nil {
		return Config{}, err
	}
	configuration.SessionCookieSecure = secureCookie
	if configuration.CSRFCookieDomain == "" && configuration.Environment == "production" {
		if publicURL, parseErr := url.Parse(configuration.PublicURL); parseErr == nil {
			configuration.CSRFCookieDomain = publicURL.Hostname()
		}
	}

	if raw := value(lookup, "ATLAS_SESSION_TTL"); raw != "" {
		configuration.SessionTTL, err = time.ParseDuration(raw)
		if err != nil || configuration.SessionTTL <= 0 {
			return Config{}, errors.New("ATLAS_SESSION_TTL must be a positive duration")
		}
	}

	if raw := value(lookup, "ATLAS_OPENROUTER_MAX_COST_USD"); raw != "" {
		configuration.OpenRouterMaxCostUSD, err = strconv.ParseFloat(raw, 64)
		if err != nil || configuration.OpenRouterMaxCostUSD < 0 {
			return Config{}, errors.New("ATLAS_OPENROUTER_MAX_COST_USD must be a non-negative number")
		}
	}

	origins := value(lookup, "ATLAS_CORS_ORIGINS")
	if origins == "" {
		origins = configuration.PublicURL
	}
	for _, origin := range strings.Split(origins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			configuration.AllowedOrigins = append(configuration.AllowedOrigins, origin)
		}
	}

	if err := configuration.Validate(); err != nil {
		return Config{}, err
	}
	return configuration, nil
}

// Validate rejects unsafe or internally inconsistent settings.
func (configuration Config) Validate() error {
	if strings.TrimSpace(configuration.DatabaseURL) == "" {
		return errors.New("ATLAS_DATABASE_URL is required")
	}
	if err := validateHTTPURL("ATLAS_PUBLIC_URL", configuration.PublicURL); err != nil {
		return err
	}
	if err := validateHTTPURL("ATLAS_API_URL", configuration.APIURL); err != nil {
		return err
	}
	if err := validateHTTPURL("ATLAS_OPENROUTER_BASE_URL", configuration.OpenRouterBaseURL); err != nil {
		return err
	}
	if configuration.LogLevel != "debug" && configuration.LogLevel != "info" && configuration.LogLevel != "warn" && configuration.LogLevel != "error" {
		return errors.New("ATLAS_LOG_LEVEL must be debug, info, warn, or error")
	}
	if configuration.CookieSameSite != "Lax" && configuration.CookieSameSite != "Strict" {
		return errors.New("ATLAS_COOKIE_SAME_SITE must be Lax or Strict")
	}
	if configuration.Environment == "production" && !configuration.SessionCookieSecure {
		return errors.New("ATLAS_COOKIE_SECURE must be true in production")
	}
	if configuration.CSRFCookieDomain != "" {
		if strings.ContainsAny(configuration.CSRFCookieDomain, "/:@ \t\r\n") {
			return errors.New("ATLAS_CSRF_COOKIE_DOMAIN must be a hostname")
		}
		publicURL, publicErr := url.Parse(configuration.PublicURL)
		apiURL, apiErr := url.Parse(configuration.APIURL)
		if publicErr != nil || apiErr != nil ||
			!hostBelongsToDomain(publicURL.Hostname(), configuration.CSRFCookieDomain) ||
			!hostBelongsToDomain(apiURL.Hostname(), configuration.CSRFCookieDomain) {
			return errors.New("ATLAS_CSRF_COOKIE_DOMAIN must contain both the public and API hosts")
		}
	}
	if configuration.OpenRouterAPIKey != "" && configuration.OpenRouterModel == "" {
		return errors.New("ATLAS_OPENROUTER_MODEL is required when ATLAS_OPENROUTER_API_KEY is configured")
	}
	for _, origin := range configuration.AllowedOrigins {
		if err := validateHTTPURL("ATLAS_CORS_ORIGINS", origin); err != nil {
			return err
		}
	}
	return nil
}

func hostBelongsToDomain(host string, domain string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	return host == domain || strings.HasSuffix(host, "."+domain)
}

func validateHTTPURL(setting string, raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be an absolute HTTP(S) URL", setting)
	}
	return nil
}

func value(lookup LookupFunc, key string) string {
	result, _ := lookup(key)
	return strings.TrimSpace(result)
}

func valueOrDefault(lookup LookupFunc, key string, fallback string) string {
	result := value(lookup, key)
	if result == "" {
		return fallback
	}
	return result
}

func boolValue(lookup LookupFunc, key string, fallback bool) (bool, error) {
	raw := value(lookup, key)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return parsed, nil
}
