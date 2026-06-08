package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAllowWithinLimit(t *testing.T) {
	l := New(3, time.Minute)
	require.True(t, l.Allow("k"))
	require.True(t, l.Allow("k"))
	require.True(t, l.Allow("k"))
	require.False(t, l.Allow("k"), "4th hit exceeds the limit of 3")
}

func TestKeysAreIndependent(t *testing.T) {
	l := New(1, time.Minute)
	require.True(t, l.Allow("a"))
	require.True(t, l.Allow("b"))
	require.False(t, l.Allow("a"))
}

func TestResetClears(t *testing.T) {
	l := New(1, time.Minute)
	require.True(t, l.Allow("k"))
	require.False(t, l.Allow("k"))
	l.Reset("k")
	require.True(t, l.Allow("k"))
}

func TestWindowExpiry(t *testing.T) {
	l := New(1, time.Minute)
	now := time.Now()
	l.now = func() time.Time { return now }

	require.True(t, l.Allow("k"))
	require.False(t, l.Allow("k"))

	now = now.Add(2 * time.Minute)
	require.True(t, l.Allow("k"), "window should have reset")
}
