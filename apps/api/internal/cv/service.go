package cv

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	Nome       string `json:"nome"`
	Titulo     string `json:"titulo"`
	TituloAlvo string `json:"titulo_alvo"`
	Resumo     string `json:"resumo"`
}

// SaveProfile validates and upserts the singleton profile.
func (s *Service) SaveProfile(ctx context.Context, in ProfileInput) (*Profile, error) {
	p := &Profile{
		Nome:       strings.TrimSpace(in.Nome),
		Titulo:     strings.TrimSpace(in.Titulo),
		TituloAlvo: strings.TrimSpace(in.TituloAlvo),
		Resumo:     strings.TrimSpace(in.Resumo),
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
	if err := s.repo.SaveProfile(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.GetProfile(ctx)
}
