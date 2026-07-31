package identity_test

import (
	"testing"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashVerifiesCorrectPassword(t *testing.T) {
	hash, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)

	verified, err := identity.VerifyPassword("a strong and memorable passphrase", hash)

	require.NoError(t, err)
	require.True(t, verified)
}

func TestPasswordHashRejectsWrongPassword(t *testing.T) {
	hash, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)

	verified, err := identity.VerifyPassword("not the password", hash)

	require.NoError(t, err)
	require.False(t, verified)
}

func TestPasswordHashUsesUniqueSalt(t *testing.T) {
	first, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)
	second, err := identity.HashPassword("a strong and memorable passphrase")
	require.NoError(t, err)

	require.NotEqual(t, first, second)
}

func TestPasswordVerificationRejectsMalformedHash(t *testing.T) {
	verified, err := identity.VerifyPassword("a strong and memorable passphrase", "not-a-password-hash")

	require.Error(t, err)
	require.False(t, verified)
}

func TestPasswordHashRejectsShortPassword(t *testing.T) {
	_, err := identity.HashPassword("too short")

	require.ErrorContains(t, err, "at least 12 characters")
}
