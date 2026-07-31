package runtime_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/config"
	appRuntime "github.com/prabhavalabs/atlas/internal/platform/runtime"
	"github.com/stretchr/testify/require"
)

func TestOpenWebRequiresBuiltIndex(t *testing.T) {
	root := t.TempDir()

	_, err := appRuntime.OpenWeb(root)

	require.ErrorContains(t, err, "index.html")
}

func TestOpenWebReturnsRootedFilesystem(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("Atlas"), 0o600))

	web, err := appRuntime.OpenWeb(root)

	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(root, "index.html"))
	require.NoError(t, err)
	require.Equal(t, "Atlas", string(content))
	index, err := web.Open("index.html")
	require.NoError(t, err)
	require.NoError(t, index.Close())
}

func TestCreateAdminRejectsInvalidEmailBeforeOpeningDatabase(t *testing.T) {
	_, err := appRuntime.CreateAdmin(
		context.Background(),
		config.Config{},
		"not-an-email",
		"Atlas Administrator",
		"correct horse battery staple",
		identity.RoleAdministrator,
	)

	require.ErrorContains(t, err, "email")
}

func TestCreateAdminRejectsBlankDisplayNameBeforeOpeningDatabase(t *testing.T) {
	_, err := appRuntime.CreateAdmin(
		context.Background(),
		config.Config{},
		"admin@example.org",
		"  ",
		"correct horse battery staple",
		identity.RoleAdministrator,
	)

	require.ErrorContains(t, err, "display name")
}
