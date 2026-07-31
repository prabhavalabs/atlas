// Package runtime owns the Atlas process lifecycle and operator commands.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/app"
	"github.com/prabhavalabs/atlas/internal/platform/config"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/prabhavalabs/atlas/internal/source/fixture"
)

const gracefulShutdownTimeout = 20 * time.Second

// OpenWeb returns a filesystem only when the Vite build is deployable.
func OpenWeb(root string) (fs.FS, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("ATLAS_WEB_ROOT is required")
	}
	web := os.DirFS(root)
	info, err := fs.Stat(web, "index.html")
	if err != nil || info.IsDir() {
		return nil, fmt.Errorf("web build index.html is unavailable under %s", root)
	}
	return web, nil
}

// Serve migrates dependencies, starts HTTP, and shuts down gracefully on context cancellation.
func Serve(ctx context.Context, configuration config.Config, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	if err := database.Migrate(ctx, configuration.DatabaseURL); err != nil {
		return err
	}
	pool, err := database.Open(ctx, configuration.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	web, err := OpenWeb(configuration.WebRoot)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              configuration.HTTPAddress,
		Handler:           app.NewHandler(app.HandlerOptions{Pool: pool, Config: configuration, Web: web}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("atlas HTTP server started", "address", configuration.HTTPAddress)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		logger.Info("atlas HTTP server stopped")
		return nil
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	}
}

// Migrate applies all embedded schema migrations.
func Migrate(ctx context.Context, configuration config.Config) error {
	return database.Migrate(ctx, configuration.DatabaseURL)
}

// ImportFixture imports one deterministic development fixture.
func ImportFixture(ctx context.Context, configuration config.Config, reader io.Reader) (fixture.Result, error) {
	if err := database.Migrate(ctx, configuration.DatabaseURL); err != nil {
		return fixture.Result{}, err
	}
	pool, err := database.Open(ctx, configuration.DatabaseURL)
	if err != nil {
		return fixture.Result{}, err
	}
	defer pool.Close()
	return fixture.Import(ctx, pool, reader)
}

// CreateAdmin creates a local administrator identity using the same password policy as login.
func CreateAdmin(ctx context.Context, configuration config.Config, email string, displayName string, password string, role identity.Role) (identity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Name != "" || parsedEmail.Address != email {
		return identity.User{}, errors.New("email must be a valid address")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return identity.User{}, errors.New("display name is required")
	}
	if role != identity.RoleAdministrator && role != identity.RoleEditor && role != identity.RoleViewer {
		return identity.User{}, errors.New("role must be administrator, editor, or viewer")
	}
	passwordHash, err := identity.HashPassword(password)
	if err != nil {
		return identity.User{}, err
	}
	if err := database.Migrate(ctx, configuration.DatabaseURL); err != nil {
		return identity.User{}, err
	}
	pool, err := database.Open(ctx, configuration.DatabaseURL)
	if err != nil {
		return identity.User{}, err
	}
	defer pool.Close()
	return identity.NewRepository(pool).CreateUser(ctx, email, displayName, passwordHash, role)
}
