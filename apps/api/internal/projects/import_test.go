package projects

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/httpx"
)

// localFetcher allows private IPs so tests can hit an httptest server on loopback.
func localFetcher() *httpx.Fetcher {
	return httpx.NewFetcher(httpx.Policy{AllowPrivateIPs: true, MaxBodyBytes: 1 << 20})
}

func TestParseGitHubRepo(t *testing.T) {
	cases := []struct {
		in          string
		owner, repo string
		ok          bool
	}{
		{"https://github.com/lumni/mirante", "lumni", "mirante", true},
		{"https://github.com/lumni/mirante.git", "lumni", "mirante", true},
		{"https://github.com/lumni/mirante/", "lumni", "mirante", true},
		{"https://github.com/lumni/mirante/tree/main", "lumni", "mirante", true},
		{"http://github.com/lumni/mirante", "lumni", "mirante", true},
		{"github.com/lumni/mirante", "lumni", "mirante", true},
		{"www.github.com/lumni/mirante", "lumni", "mirante", true},
		{"git@github.com:lumni/mirante.git", "lumni", "mirante", true},
		{"  https://github.com/lumni/mirante  ", "lumni", "mirante", true},
		{"https://gitlab.com/lumni/mirante", "", "", false},
		{"https://github.com/lumni", "", "", false},
		{"not-a-url", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		owner, repo, ok := parseGitHubRepo(c.in)
		require.Equalf(t, c.ok, ok, "ok for %q", c.in)
		if c.ok {
			require.Equalf(t, c.owner, owner, "owner for %q", c.in)
			require.Equalf(t, c.repo, repo, "repo for %q", c.in)
		}
	}
}

func TestHumanizeName(t *testing.T) {
	require.Equal(t, "Mirante", humanizeName("mirante"))
	require.Equal(t, "My Cool App", humanizeName("my-cool_app"))
	require.Equal(t, "Lumni Console", humanizeName("lumni.console"))
	require.Equal(t, "API Gateway", humanizeName("API-gateway")) // acronym preserved
	require.Equal(t, "", humanizeName(""))
}

func TestBuildTags(t *testing.T) {
	// Language first; "go" topic de-duped case-insensitively against the language.
	require.Equal(t, []string{"Go", "cli", "self-hosted"},
		buildTags("Go", []string{"go", "cli", "self-hosted"}))
	require.Equal(t, []string{"TypeScript"}, buildTags("TypeScript", nil))
	require.Empty(t, buildTags("", nil))
}

const ghJSON = `{
  "name": "Mirante-App",
  "description": "  Central pessoal de comando  ",
  "html_url": "https://github.com/lumni/Mirante-App",
  "language": "Go",
  "topics": ["go", "sveltekit", "self-hosted"],
  "archived": false
}`

func TestImportDraftSuccess(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, ghJSON)
	}))
	t.Cleanup(srv.Close)

	svc := &Service{fetcher: localFetcher(), githubAPI: srv.URL}
	d, err := svc.ImportDraft(context.Background(), "https://github.com/lumni/Mirante-App.git")
	require.NoError(t, err)
	require.Equal(t, "/repos/lumni/Mirante-App", gotPath)
	require.Equal(t, "Mirante App", d.Nome)     // humanized from "Mirante-App"
	require.Equal(t, "Mirante-App", d.Codinome) // raw slug from the API
	require.Equal(t, "Central pessoal de comando", d.Descricao)
	require.Equal(t, "https://github.com/lumni/Mirante-App", d.Repo)
	require.Equal(t, []string{"Go", "sveltekit", "self-hosted"}, d.Tags)
	require.Equal(t, Status(""), d.Status)
}

func TestImportDraftArchived(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"name":"old","html_url":"https://github.com/x/old","archived":true}`)
	}))
	t.Cleanup(srv.Close)

	svc := &Service{fetcher: localFetcher(), githubAPI: srv.URL}
	d, err := svc.ImportDraft(context.Background(), "github.com/x/old")
	require.NoError(t, err)
	require.Equal(t, StatusArquivado, d.Status)
}

func TestImportDraftNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"Not Found"}`)
	}))
	t.Cleanup(srv.Close)

	svc := &Service{fetcher: localFetcher(), githubAPI: srv.URL}
	_, err := svc.ImportDraft(context.Background(), "https://github.com/x/missing")
	require.ErrorIs(t, err, ErrImportFailed)
}

func TestImportDraftBadURLAndNoFetcher(t *testing.T) {
	// A non-GitHub URL is rejected before the fetcher is consulted.
	withFetcher := &Service{fetcher: localFetcher(), githubAPI: defaultGitHubAPI}
	_, err := withFetcher.ImportDraft(context.Background(), "https://gitlab.com/x/y")
	require.ErrorIs(t, err, ErrInvalid)

	// A valid GitHub URL with no fetcher wired reports the feature is unavailable.
	noFetcher := &Service{githubAPI: defaultGitHubAPI}
	_, err = noFetcher.ImportDraft(context.Background(), "https://github.com/x/y")
	require.ErrorIs(t, err, ErrImportUnavailable)
}
