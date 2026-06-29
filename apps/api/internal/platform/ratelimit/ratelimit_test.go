package ratelimit

import (
	"fmt"
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

func TestCardinalityCapBoundsMap(t *testing.T) {
	l := New(10, time.Minute)
	l.maxKeys = 50
	now := time.Now()
	l.now = func() time.Time { return now }

	// Fill the map to its cap with distinct, still-live keys.
	for i := 0; i < l.maxKeys; i++ {
		require.True(t, l.Allow(fmt.Sprintf("k%d", i)))
	}
	require.Len(t, l.hits, l.maxKeys)

	// Far more distinct keys arrive: the map stays bounded (eviction makes room),
	// and a legitimate new key is still admitted (never locked out).
	for i := 0; i < 10*l.maxKeys; i++ {
		require.True(t, l.Allow(fmt.Sprintf("flood%d", i)), "new keys must still be admitted")
		require.LessOrEqual(t, len(l.hits), l.maxKeys, "map must never exceed the cap")
	}

	// An already-tracked, still-live key keeps its counter and limit.
	l.hits["sticky"] = &counter{count: l.max, resetAt: now.Add(time.Minute)}
	require.False(t, l.Allow("sticky"), "a key at its limit is still throttled")
}
