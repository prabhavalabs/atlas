// Package migrations embeds Atlas database migrations in the application image.
package migrations

import "embed"

// Files contains all Goose migrations at build time.
//
//go:embed *.sql
var Files embed.FS
