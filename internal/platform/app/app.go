// Package app assembles the Atlas modular monolith from narrow domain modules.
package app

import (
	"context"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prabhavalabs/atlas/internal/adminapi"
	"github.com/prabhavalabs/atlas/internal/event"
	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/buildinfo"
	"github.com/prabhavalabs/atlas/internal/platform/config"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/prabhavalabs/atlas/internal/platform/httpserver"
	"github.com/prabhavalabs/atlas/internal/publicapi"
)

// HandlerOptions contains the process dependencies needed by the HTTP boundary.
type HandlerOptions struct {
	Pool   *pgxpool.Pool
	Config config.Config
	Web    fs.FS
	Now    func() time.Time
}

// NewHandler assembles public, administrator, health, and web routes.
func NewHandler(options HandlerOptions) http.Handler {
	events := event.NewRepository(options.Pool)
	identities := identity.NewRepository(options.Pool)
	return httpserver.New(httpserver.Options{
		Build:          buildinfo.Current(),
		Ready:          func(ctx context.Context) error { return database.Ready(ctx, options.Pool) },
		AllowedOrigins: options.Config.AllowedOrigins,
		Now:            options.Now,
		Web:            options.Web,
		APIOrigin:      options.Config.APIURL,
		MetricsToken:   options.Config.MetricsToken,
		Register: func(router chi.Router) {
			publicapi.Register(router, events, options.Now)
			adminapi.Register(router, identities, events, adminapi.Options{
				SecureCookie:     options.Config.SessionCookieSecure,
				CSRFCookieDomain: options.Config.CSRFCookieDomain,
				SessionTTL:       options.Config.SessionTTL,
				Now:              options.Now,
			})
		},
	})
}
