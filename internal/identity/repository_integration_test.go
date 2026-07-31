//go:build integration

package identity_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/database"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreatesAuthenticatesAndRevokesSession(t *testing.T) {
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))
	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE audit_log, sessions, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	repository := identity.NewRepository(pool)
	passwordHash, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)
	user, err := repository.CreateUser(ctx, "editor@example.org", "Atlas Editor", passwordHash, identity.RoleEditor)
	require.NoError(t, err)

	tokens, err := identity.NewSessionTokens()
	require.NoError(t, err)
	now := time.Date(2026, 7, 31, 20, 0, 0, 0, time.UTC)
	session, err := repository.CreateSession(ctx, user.ID, tokens, now.Add(time.Hour), now)
	require.NoError(t, err)

	authenticated, err := repository.AuthenticateSession(ctx, tokens.SessionToken, now.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, user.ID, authenticated.User.ID)
	require.Equal(t, identity.RoleEditor, authenticated.User.Role)
	require.True(t, identity.MatchesSecret(tokens.CSRFToken, authenticated.CSRFHash))

	_, err = repository.AuthenticateSession(ctx, "wrong-session", now.Add(time.Minute))
	require.True(t, errors.Is(err, identity.ErrInvalidSession))

	require.NoError(t, repository.RevokeSession(ctx, session.ID, now.Add(2*time.Minute)))
	_, err = repository.AuthenticateSession(ctx, tokens.SessionToken, now.Add(3*time.Minute))
	require.True(t, errors.Is(err, identity.ErrInvalidSession))
}

func TestRepositoryRejectsExpiredSession(t *testing.T) {
	databaseURL := os.Getenv("ATLAS_TEST_DATABASE_URL")
	require.NotEmpty(t, databaseURL)
	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx, databaseURL))
	pool, err := database.Open(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE audit_log, sessions, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	repository := identity.NewRepository(pool)
	passwordHash, err := identity.HashPassword("a different strong passphrase")
	require.NoError(t, err)
	user, err := repository.CreateUser(ctx, "admin@example.org", "Atlas Admin", passwordHash, identity.RoleAdministrator)
	require.NoError(t, err)
	tokens, err := identity.NewSessionTokens()
	require.NoError(t, err)
	now := time.Date(2026, 7, 31, 20, 0, 0, 0, time.UTC)
	_, err = repository.CreateSession(ctx, user.ID, tokens, now.Add(time.Minute), now)
	require.NoError(t, err)

	_, err = repository.AuthenticateSession(ctx, tokens.SessionToken, now.Add(2*time.Minute))

	require.True(t, errors.Is(err, identity.ErrInvalidSession))
}
