package projects

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lumni/mirante/internal/platform/httpx"
	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/validate"
)

// Errors.
var (
	// ErrInvalid wraps validation failures (mapped to HTTP 400).
	ErrInvalid = errors.New("invalid input")
	// ErrImportUnavailable means GitHub import is off (no fetcher wired).
	ErrImportUnavailable = errors.New("import unavailable")
	// ErrImportFailed wraps a failed GitHub import; the wrapped reason is
	// user-facing, so its message reads as a pt-BR sentence prefix.
	ErrImportFailed = errors.New("falha ao importar do GitHub")
)

const defaultGitHubAPI = "https://api.github.com"

// Service holds project use cases. The GitHub fetcher is optional (nil-safe):
// without it, ImportDraft returns ErrImportUnavailable; all CRUD works regardless.
type Service struct {
	repo      Repository
	fetcher   *httpx.Fetcher
	githubAPI string
}

// NewService builds the projects service. fetcher may be nil (GitHub import off).
func NewService(repo Repository, fetcher *httpx.Fetcher) *Service {
	return &Service{repo: repo, fetcher: fetcher, githubAPI: defaultGitHubAPI}
}

// CreateInput is the payload for creating a project.
type CreateInput struct {
	Nome         string     `json:"nome"`
	Codinome     string     `json:"codinome"`
	Descricao    string     `json:"descricao"`
	Repo         string     `json:"repo"`
	Status       Status     `json:"status"`
	Visibilidade Visibility `json:"visibilidade"`
	Tags         []string   `json:"tags"`
}

// UpdateInput is a partial update; nil fields are left unchanged.
type UpdateInput struct {
	Nome         *string     `json:"nome"`
	Codinome     *string     `json:"codinome"`
	Descricao    *string     `json:"descricao"`
	Repo         *string     `json:"repo"`
	Status       *Status     `json:"status"`
	Visibilidade *Visibility `json:"visibilidade"`
	Tags         *[]string   `json:"tags"`
}

// LinkInput is the payload for attaching a link.
type LinkInput struct {
	Label     string `json:"label" validate:"required,max=80"`
	URL       string `json:"url" validate:"required,url"`
	Kind      string `json:"kind" validate:"omitempty,oneof=prod staging repo docs design other"`
	SortOrder int    `json:"sort_order"`
}

// Get returns a project with links and tags.
func (s *Service) Get(ctx context.Context, id ID) (*Project, error) {
	return s.repo.Get(ctx, id)
}

// List returns projects (optionally filtered by status).
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Project, error) {
	return s.repo.List(ctx, f)
}

// Create validates and persists a new project.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Project, error) {
	p := &Project{
		ID:           ID(idgen.New()),
		Nome:         strings.TrimSpace(in.Nome),
		Codinome:     strings.TrimSpace(in.Codinome),
		Descricao:    in.Descricao,
		Repo:         strings.TrimSpace(in.Repo),
		Status:       in.Status,
		Visibilidade: in.Visibilidade,
	}
	if p.Status == "" {
		p.Status = StatusIdeia
	}
	if p.Visibilidade == "" {
		p.Visibilidade = VisPessoal
	}
	if err := validateProject(p); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	if len(in.Tags) > 0 {
		if err := s.repo.SetTags(ctx, p.ID, in.Tags); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, p.ID)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id ID, in UpdateInput) (*Project, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Nome != nil {
		p.Nome = strings.TrimSpace(*in.Nome)
	}
	if in.Codinome != nil {
		p.Codinome = strings.TrimSpace(*in.Codinome)
	}
	if in.Descricao != nil {
		p.Descricao = *in.Descricao
	}
	if in.Repo != nil {
		p.Repo = strings.TrimSpace(*in.Repo)
	}
	if in.Status != nil {
		p.Status = *in.Status
	}
	if in.Visibilidade != nil {
		p.Visibilidade = *in.Visibilidade
	}
	if err := validateProject(p); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	if in.Tags != nil {
		if err := s.repo.SetTags(ctx, id, *in.Tags); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, id)
}

// Delete hard-deletes a project (cascading links, tags, and — via FK — its
// monitor services). Use Update(status=arquivado) for a soft archive.
func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

// AddLink validates and attaches a link, returning the updated project.
func (s *Service) AddLink(ctx context.Context, projectID ID, in LinkInput) (*Project, error) {
	if err := validate.Struct(in); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	// Ownership check (Get is user-scoped): refuse to attach a link to a project
	// the caller doesn't own, instead of leaving an orphan row.
	if _, err := s.repo.Get(ctx, projectID); err != nil {
		return nil, err
	}
	kind := in.Kind
	if kind == "" {
		kind = "other"
	}
	l := &Link{
		ID:        ID(idgen.New()),
		ProjectID: projectID,
		Label:     in.Label,
		URL:       in.URL,
		Kind:      kind,
		SortOrder: in.SortOrder,
	}
	if err := s.repo.AddLink(ctx, l); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, projectID)
}

// RemoveLink detaches a link.
func (s *Service) RemoveLink(ctx context.Context, projectID, linkID ID) error {
	return s.repo.RemoveLink(ctx, projectID, linkID)
}

func validateProject(p *Project) error {
	if n := strings.TrimSpace(p.Nome); n == "" || len(n) > 200 {
		return fmt.Errorf("%w: nome is required (max 200)", ErrInvalid)
	}
	if err := validate.Var(string(p.Status), "oneof=ideia ativo pausado no_ar arquivado"); err != nil {
		return fmt.Errorf("%w: status", ErrInvalid)
	}
	if err := validate.Var(string(p.Visibilidade), "oneof=pessoal lumni cliente"); err != nil {
		return fmt.Errorf("%w: visibilidade", ErrInvalid)
	}
	if p.Repo != "" {
		if err := validate.Var(p.Repo, "url"); err != nil {
			return fmt.Errorf("%w: repo must be a URL", ErrInvalid)
		}
	}
	return nil
}
