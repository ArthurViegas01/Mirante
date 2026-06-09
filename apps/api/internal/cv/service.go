package cv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lumni/mirante/internal/skills"
)

// ErrInvalid wraps validation failures (mapped to HTTP 400).
var ErrInvalid = errors.New("invalid input")

// Service holds CV use cases.
type Service struct{ repo Repository }

// NewService builds the cv service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// GetProfile returns the master profile (empty if never saved).
func (s *Service) GetProfile(ctx context.Context) (*Profile, error) {
	return s.repo.GetProfile(ctx)
}

// ProfileInput is the payload for saving the profile.
type ProfileInput struct {
	Nome       string   `json:"nome"`
	Titulo     string   `json:"titulo"`
	TituloAlvo string   `json:"titulo_alvo"`
	Resumo     string   `json:"resumo"`
	Skills     []string `json:"skills"`
}

// SaveProfile validates and upserts the singleton profile (including its master
// skills, canonicalized via the skills kernel).
func (s *Service) SaveProfile(ctx context.Context, in ProfileInput) (*Profile, error) {
	p := &Profile{
		Nome:       strings.TrimSpace(in.Nome),
		Titulo:     strings.TrimSpace(in.Titulo),
		TituloAlvo: strings.TrimSpace(in.TituloAlvo),
		Resumo:     strings.TrimSpace(in.Resumo),
		Skills:     normalizeSkills(in.Skills),
	}
	for _, f := range []struct{ value, name string }{
		{p.Nome, "nome"}, {p.Titulo, "titulo"}, {p.TituloAlvo, "titulo_alvo"},
	} {
		if len([]rune(f.value)) > 120 {
			return nil, fmt.Errorf("%w: %s muito longo (max 120)", ErrInvalid, f.name)
		}
	}
	if len([]rune(p.Resumo)) > 2000 {
		return nil, fmt.Errorf("%w: resumo muito longo (max 2000)", ErrInvalid)
	}
	if len(p.Skills) > 100 {
		return nil, fmt.Errorf("%w: muitas skills (max 100)", ErrInvalid)
	}
	if err := s.repo.SaveProfile(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.GetProfile(ctx)
}

// normalizeSkills trims, canonicalizes (via skills.Normalize when recognized),
// and de-duplicates the master skill list.
func normalizeSkills(raw []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, sk := range raw {
		sk = strings.TrimSpace(sk)
		if sk == "" {
			continue
		}
		if canon, ok := skills.Normalize(sk); ok {
			sk = canon
		}
		key := strings.ToLower(sk)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, sk)
	}
	return out
}
