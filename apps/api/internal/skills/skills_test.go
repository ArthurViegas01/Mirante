package skills

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"golang":      "Go",
		"Go":          "Go",
		"GoLang":      "Go", // case-insensitive
		"  react  ":   "React",
		"js":          "JavaScript",
		"ts":          "TypeScript",
		"postgres":    "PostgreSQL",
		"psql":        "PostgreSQL",
		"k8s":         "Kubernetes",
		"spring boot": "Spring",
		"dotnet":      ".NET",
	}
	for raw, want := range cases {
		got, ok := Normalize(raw)
		require.Truef(t, ok, "Normalize(%q) should resolve", raw)
		require.Equalf(t, want, got, "Normalize(%q)", raw)
	}

	_, ok := Normalize("definitely-not-a-skill")
	require.False(t, ok)
	_, ok = Normalize("")
	require.False(t, ok)
}

func TestMatchOrderAndDedupe(t *testing.T) {
	text := "Vaga para dev Go com React, PostgreSQL, Docker e Kubernetes (k8s)."
	got := Match(text)
	// Returned in catalog order; Kubernetes counted once despite "Kubernetes"+"k8s".
	require.Equal(t, []string{"Go", "React", "PostgreSQL", "Docker", "Kubernetes"}, got)
}

func TestMatchTokenBoundaries(t *testing.T) {
	// "go" inside another word must not match the Go language.
	require.Empty(t, Match("categoria de jogos, alongamento e congo"))
	// "sql" inside PostgreSQL/MySQL must not surface a bare SQL.
	require.NotContains(t, Match("usei PostgreSQL no projeto"), "SQL")
	require.Contains(t, Match("usei PostgreSQL no projeto"), "PostgreSQL")
	// Standalone token does match.
	require.Contains(t, Match("escrevi SQL puro"), "SQL")
}

func TestMatchSpecialCharsAndMultiword(t *testing.T) {
	got := Match("Experiência com C#, C++ e .NET; também Node.js e Spring Boot.")
	require.Contains(t, got, "C#")
	require.Contains(t, got, "C++")
	require.Contains(t, got, ".NET")
	require.Contains(t, got, "Node.js")
	require.Contains(t, got, "Spring")
}

func TestMatchAliases(t *testing.T) {
	got := Match("stack: golang, tailwindcss, turso e github actions")
	require.Contains(t, got, "Go")
	require.Contains(t, got, "Tailwind CSS")
	require.Contains(t, got, "libSQL")
	require.Contains(t, got, "GitHub Actions")
}

func TestCatalogIntegrity(t *testing.T) {
	seenCanon := map[string]bool{}
	seenKey := map[string]string{} // lowercased alias/canonical -> owning canonical
	for _, s := range catalog {
		require.NotEmpty(t, s.Canonical)
		require.Falsef(t, seenCanon[s.Canonical], "duplicate canonical %q", s.Canonical)
		seenCanon[s.Canonical] = true

		keys := append([]string{s.Canonical}, s.Aliases...)
		for _, k := range keys {
			lk := strings.ToLower(k)
			if owner, dup := seenKey[lk]; dup {
				t.Fatalf("alias/canonical %q (skill %q) collides with skill %q", k, s.Canonical, owner)
			}
			seenKey[lk] = s.Canonical
		}
		// Related must point at real canonicals.
		for _, r := range s.Related {
			_, ok := Get(r)
			require.Truef(t, ok, "skill %q relates to unknown %q", s.Canonical, r)
		}
	}
}

func TestAccessors(t *testing.T) {
	require.NotEmpty(t, All())
	s, ok := Get("SvelteKit")
	require.True(t, ok)
	require.Equal(t, CatFramework, s.Category)
	require.Contains(t, s.Related, "Svelte")

	groups := ByCategory()
	require.NotEmpty(t, groups[CatLinguagem])
	require.NotEmpty(t, Categories())
}
