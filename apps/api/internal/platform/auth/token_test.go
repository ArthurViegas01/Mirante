package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashTokenDeterministic(t *testing.T) {
	require.Equal(t, hashToken("abc"), hashToken("abc"))
	require.NotEqual(t, hashToken("abc"), hashToken("abd"))
}

func TestNewTokenUnique(t *testing.T) {
	a, err := newToken()
	require.NoError(t, err)
	b, err := newToken()
	require.NoError(t, err)
	require.NotEqual(t, a, b)
	require.GreaterOrEqual(t, len(a), 40, "256-bit base64url token should be ~43 chars")
}
