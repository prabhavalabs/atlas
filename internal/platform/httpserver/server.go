// Package httpserver assembles Atlas HTTP routing and cross-cutting middleware.
package httpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prabhavalabs/atlas/internal/platform/buildinfo"
)

const requestIDHeader = "X-Request-ID"

// Options controls the dependency boundaries of the HTTP server.
type Options struct {
	Build          buildinfo.Info
	Ready          func(context.Context) error
	AllowedOrigins []string
	Now            func() time.Time
	Register       func(chi.Router)
	Web            fs.FS
	APIOrigin      string
	MetricsToken   string
}

// New returns the complete Atlas HTTP handler.
func New(options Options) http.Handler {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Ready == nil {
		options.Ready = func(context.Context) error { return nil }
	}

	router := chi.NewRouter()
	router.Use(recoverer)
	router.Use(requestID)
	router.Use(securityHeaders(options.APIOrigin))
	router.Use(cors(options.AllowedOrigins))
	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		if options.Web != nil && canServeWeb(request) {
			serveWeb(options.Web, writer, request)
			return
		}
		writeProblem(writer, request, http.StatusNotFound, "route_not_found", "Route not found")
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		writeProblem(writer, request, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	})

	router.Get("/health/live", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/health/ready", func(writer http.ResponseWriter, request *http.Request) {
		if err := options.Ready(request.Context()); err != nil {
			writeProblem(writer, request, http.StatusServiceUnavailable, "dependency_unavailable", "A required dependency is unavailable")
			return
		}
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ready"})
	})
	router.Get("/api/v1/meta", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]string{
			"version":     options.Build.Version,
			"commit":      options.Build.Commit,
			"buildTime":   options.Build.BuildTime,
			"generatedAt": options.Now().UTC().Format(time.RFC3339),
		})
	})
	router.Get("/metrics", metrics(options.MetricsToken))

	if options.Register != nil {
		options.Register(router)
	}
	return router
}

func metrics(expectedToken string) http.HandlerFunc {
	expectedHash := sha256.Sum256([]byte(expectedToken))
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
		providedHash := sha256.Sum256([]byte(provided))
		if expectedToken == "" || subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
			writeProblem(writer, request, http.StatusUnauthorized, "metrics_authentication_required", "Metrics authentication is required")
			return
		}
		writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("# HELP atlas_up Whether the Atlas HTTP process is running.\n# TYPE atlas_up gauge\natlas_up 1\n"))
	}
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		identifier := request.Header.Get(requestIDHeader)
		if identifier == "" || len(identifier) > 128 {
			identifier = randomID()
		}
		writer.Header().Set(requestIDHeader, identifier)
		request.Header.Set(requestIDHeader, identifier)
		next.ServeHTTP(writer, request)
	})
}

func randomID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(buffer)
}

func securityHeaders(apiURL string) func(http.Handler) http.Handler {
	connectSources := "'self'"
	if parsed, err := url.Parse(apiURL); err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		connectSources += " " + parsed.Scheme + "://" + parsed.Host
	}
	policy := "default-src 'self'; connect-src " + connectSources + " https://tiles.openfreemap.org; img-src 'self' data: blob: https://tiles.openfreemap.org; style-src 'self' 'unsafe-inline'; font-src 'self'; worker-src 'self' blob:; frame-ancestors 'none'; object-src 'none'; base-uri 'self'"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("X-Content-Type-Options", "nosniff")
			writer.Header().Set("X-Frame-Options", "DENY")
			writer.Header().Set("Referrer-Policy", "no-referrer")
			writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			writer.Header().Set("Content-Security-Policy", policy)
			next.ServeHTTP(writer, request)
		})
	}
}

func canServeWeb(request *http.Request) bool {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		return false
	}
	return !strings.HasPrefix(request.URL.Path, "/api/") &&
		!strings.HasPrefix(request.URL.Path, "/health/") &&
		request.URL.Path != "/metrics"
}

func serveWeb(web fs.FS, writer http.ResponseWriter, request *http.Request) {
	requestedPath := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
	if requestedPath != "." && requestedPath != "" {
		if info, err := fs.Stat(web, requestedPath); err == nil && !info.IsDir() {
			writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			http.FileServer(http.FS(web)).ServeHTTP(writer, request)
			return
		}
		if strings.HasPrefix(requestedPath, "assets/") {
			writeProblem(writer, request, http.StatusNotFound, "asset_not_found", "Web asset not found")
			return
		}
	}

	index, err := fs.ReadFile(web, "index.html")
	if err != nil {
		writeProblem(writer, request, http.StatusServiceUnavailable, "web_unavailable", "Web application is unavailable")
		return
	}
	writer.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(writer, request, "index.html", time.Time{}, bytes.NewReader(index))
}

func cors(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				writer.Header().Set("Access-Control-Allow-Origin", origin)
				writer.Header().Set("Access-Control-Allow-Credentials", "true")
				writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Request-ID")
				writer.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH, DELETE, OPTIONS")
				writer.Header().Add("Vary", "Origin")
			}
			if request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != "" {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recover() != nil {
				writeProblem(writer, request, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Code     string `json:"code"`
	Instance string `json:"instance"`
}

func writeProblem(writer http.ResponseWriter, request *http.Request, status int, code string, title string) {
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(problem{
		Type:     "https://atlas.prabhavalabs.com/problems/" + code,
		Title:    title,
		Status:   status,
		Code:     code,
		Instance: request.URL.Path,
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
