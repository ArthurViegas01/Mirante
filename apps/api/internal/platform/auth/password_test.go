package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndVerifyPassword(t *testing.T) {
	const pw = "correct horse battery staple"
	hash, err := HashPassword(pw)
	require.NoError(t, err)
	require.Contains(t, hash, "$argon2id$")

	ok, err := VerifyPassword(pw, hash)
	require.NoError(t, err)
	require.True(t, ok, "correct password must verify")

	ok, err = VerifyPassword("wrong password", hash)
	require.NoError(t, err)
	require.False(t, ok, "wrong password must not verify")
}

func TestHashIsSalted(t *testing.T) {
	h1, err := HashPassword("same")
	require.NoError(t, err)
	h2, err := HashPassword("same")
	require.NoError(t, err)
	require.NotEqual(t, h1, h2, "equal passwords must produce different hashes (random salt)")
}

func TestVerifyInvalidHash(t *testing.T) {
	_, err := VerifyPassword("x", "not-a-valid-hash")
	require.ErrorIs(t, err, ErrInvalidHash)
}
