package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// Draft is an unsaved project parsed from a GitHub repo URL; the UI pre-fills the
// new-project form with it. It is NOT persisted — the user reviews and saves it.
type Draft struct {
	Nome      string   `json:"nome"`
	Codinome  string   `json:"codinome"`
	Descricao string   `json:"descricao"`
	Repo      string   `json:"repo"`
	Status    Status   `json:"status"`
	Tags      []string `json:"tags"`
}

// ghRepo is the slice of the GitHub repository API response we use.
type ghRepo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	Archived    bool     `json:"archived"`
}

// ImportDraft fetches a GitHub repository's metadata (SSRF-guarded) and turns it
// into a project draft: name (humanized slug), codename (the slug), description,
// canonical repo URL and tags (primary language first, then topics). An archived
// repo is drafted as `arquivado`. The result is NOT persisted — the user reviews
// and saves it.
func (s *Service) ImportDraft(ctx context.Context, rawURL string) (*Draft, error) {
	owner, repo, ok := parseGitHubRepo(rawURL)
	if !ok {
		return nil, fmt.Errorf("%w: informe um link do GitHub (ex.: https://github.com/usuário/repo)", ErrInvalid)
	}
	if s.fetcher == nil {
		return nil, ErrImportUnavailable
	}

	apiURL := strings.TrimRight(s.githubAPI, "/") + "/repos/" + owner + "/" + repo
	res, body, err := s.fetcher.Fetch(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImportFailed, err)
	}
	switch res.StatusCode {
	case http.StatusOK:
		// proceed
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: repositório não encontrado (pode ser privado ou o link está errado)", ErrImportFailed)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("%w: limite da API do GitHub atingido — tente novamente em alguns minutos", ErrImportFailed)
	default:
		return nil, fmt.Errorf("%w: o GitHub respondeu %d", ErrImportFailed, res.StatusCode)
	}

	var gh ghRepo
	if err := json.Unmarshal(body, &gh); err != nil {
		return nil, fmt.Errorf("%w: resposta inesperada do GitHub", ErrImportFailed)
	}

	d := &Draft{
		Nome:      humanizeName(gh.Name),
		Codinome:  gh.Name,
		Descricao: strings.TrimSpace(gh.Description),
		Repo:      strings.TrimSpace(gh.HTMLURL),
		Tags:      buildTags(gh.Language, gh.Topics),
	}
	if d.Repo == "" {
		d.Repo = "https://github.com/" + owner + "/" + repo
	}
	if gh.Archived {
		d.Status = StatusArquivado
	}
	return d, nil
}

// parseGitHubRepo extracts owner/repo from the GitHub URL forms a user is likely
// to paste: https/http URLs, scheme-less host paths, a trailing ".git", deeper
// paths (…/tree/main), and the SSH form git@github.com:owner/repo.git. Only
// github.com is accepted.
func parseGitHubRepo(raw string) (owner, repo string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	// SSH form: git@github.com:owner/repo(.git)
	if strings.HasPrefix(raw, "git@") {
		_, rest, found := strings.Cut(raw, ":")
		if !found {
			return "", "", false
		}
		return splitOwnerRepo(rest)
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false
	}
	if h := strings.ToLower(u.Hostname()); h != "github.com" && h != "www.github.com" {
		return "", "", false
	}
	return splitOwnerRepo(u.Path)
}

func splitOwnerRepo(path string) (owner, repo string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	repo = strings.TrimSuffix(parts[1], ".git")
	if repo == "" {
		return "", "", false
	}
	return parts[0], repo, true
}

// humanizeName turns a repo slug into a display name: "my-cool_app" → "My Cool
// App". The first letter of each word is upper-cased; the rest is left intact so
// acronyms already in the slug survive.
func humanizeName(name string) string {
	fields := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || r == ' '
	})
	for i, f := range fields {
		r := []rune(f)
		r[0] = unicode.ToUpper(r[0])
		fields[i] = string(r)
	}
	return strings.Join(fields, " ")
}

// buildTags lists the primary language first, then the repo topics, de-duplicated
// case-insensitively and capped to a reasonable count.
func buildTags(language string, topics []string) []string {
	const maxTags = 12
	seen := make(map[string]bool)
	tags := make([]string, 0, maxTags)
	add := func(t string) {
		t = strings.TrimSpace(t)
		if t == "" || len(tags) >= maxTags {
			return
		}
		if k := strings.ToLower(t); !seen[k] {
			seen[k] = true
			tags = append(tags, t)
		}
	}
	add(language)
	for _, t := range topics {
		add(t)
	}
	return tags
}
