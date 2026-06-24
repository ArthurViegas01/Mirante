package intake

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEmail(t *testing.T) {
	raw, err := os.ReadFile("testdata/digest.eml")
	require.NoError(t, err)

	projects, err := ParseEmail(raw)
	require.NoError(t, err)
	require.Greater(t, len(projects), 5, "digest should list several projects")

	// First project, fully resolved from the real e-mail.
	first := projects[0]
	require.Equal(t, "762750", first.FonteID)
	require.Contains(t, first.Titulo, "micro-SaaS")
	require.Equal(t, "Desenvolvimento Web", first.Categoria)
	require.Equal(t, "Especialista", first.Nivel)
	require.Equal(t, 73, first.Propostas)
	require.Equal(t, 92, first.Interessados)
	require.Equal(t,
		"https://www.99freelas.com.br/project/mvp-de-micro-saas-para-profissionais-da-saude-762750",
		first.URL)
	require.Equal(t,
		"https://www.99freelas.com.br/project/bid/mvp-de-micro-saas-para-profissionais-da-saude-762750",
		first.EnviarURL)
	require.NotContains(t, first.Teaser, "Leia mais")

	// Every project carries source + dedup metadata; the "Web, Mobile & Software"
	// section header is not mistaken for a project.
	for _, p := range projects {
		require.Equal(t, Fonte99Freelas, p.Fonte)
		require.NotEmpty(t, p.FonteID, "missing id for %q", p.Titulo)
		require.Regexp(t, `^\d+$`, p.FonteID)
		require.Contains(t, p.URL, "/project/")
		require.Contains(t, p.EnviarURL, "/project/bid/")
		require.NotEqual(t, "Web, Mobile & Software", p.Titulo)
	}
}

func TestParseEmailNotMultipart(t *testing.T) {
	raw := []byte("From: x@y.com\r\nContent-Type: multipart/alternative; boundary=b\r\n\r\nbody")
	_, err := ParseEmail(raw)
	require.Error(t, err)
}
