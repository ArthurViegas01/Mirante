package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadIntakeDefaults(t *testing.T) {
	c, err := Load()
	require.NoError(t, err)
	require.Equal(t, "imap.gmail.com", c.IntakeIMAPHost)
	require.Equal(t, 993, c.IntakeIMAPPort)
	require.Equal(t, "INBOX", c.IntakeIMAPMailbox)
	require.Equal(t, "99freelas.com.br", c.IntakeIMAPFrom)
	require.Equal(t, 15*time.Minute, c.IntakePollInterval)
	require.Equal(t, 60, c.IntakeMinScore)
	// No credentials → intake stays off.
	require.False(t, c.IntakeEnabled())
}

func TestLoadIntakeFromEnv(t *testing.T) {
	t.Setenv("INTAKE_IMAP_USERNAME", "arthur@gmail.com")
	t.Setenv("INTAKE_IMAP_PASSWORD", "app-password-16")
	t.Setenv("INTAKE_IMAP_PORT", "1993")
	t.Setenv("INTAKE_POLL_INTERVAL", "5m")
	t.Setenv("INTAKE_MIN_SCORE", "70")

	c, err := Load()
	require.NoError(t, err)
	require.True(t, c.IntakeEnabled())
	require.Equal(t, "arthur@gmail.com", c.IntakeIMAPUsername)
	require.Equal(t, 1993, c.IntakeIMAPPort)
	require.Equal(t, 5*time.Minute, c.IntakePollInterval)
	require.Equal(t, 70, c.IntakeMinScore)
}

func TestLoadIntakeInvalidInterval(t *testing.T) {
	t.Setenv("INTAKE_POLL_INTERVAL", "nope")
	_, err := Load()
	require.Error(t, err)
}
