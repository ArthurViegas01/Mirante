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

// GetProfile returns the full master CV (empty if never saved).
func (s *Service) GetProfile(ctx context.Context) (*Profile, error) {
	return s.repo.GetProfile(ctx)
}

// ProfileInput is a partial identity update (PUT /api/profile). Skills is
// optional: nil leaves the master skills unchanged (so the quick header edit on
// /vagas doesn't wipe them); experiences/education are always preserved here.
type ProfileInput struct {
	Nome       string    `json:"nome"`
	Titulo     string    `json:"titulo"`
	TituloAlvo string    `json:"titulo_alvo"`
	Resumo     string    `json:"resumo"`
	Skills     *[]string `json:"skills"`
}

// SaveProfile patches the identity (and skills if provided) on top of the current
// CV, preserving everything else.
func (s *Service) SaveProfile(ctx context.Context, in ProfileInput) (*Profile, error) {
	cur, err := s.repo.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	cur.Nome = strings.TrimSpace(in.Nome)
	cur.Titulo = strings.TrimSpace(in.Titulo)
	cur.TituloAlvo = strings.TrimSpace(in.TituloAlvo)
	cur.Resumo = strings.TrimSpace(in.Resumo)
	if in.Skills != nil {
		cur.Skills = normalizeSkills(*in.Skills)
	}
	if err := validateProfile(cur); err != nil {
		return nil, err
	}
	if err := s.repo.SaveCV(ctx, cur); err != nil {
		return nil, err
	}
	return s.repo.GetProfile(ctx)
}

// CVInput is the full master CV (PUT /api/cv).
type CVInput struct {
	Nome        string            `json:"nome"`
	Titulo      string            `json:"titulo"`
	TituloAlvo  string            `json:"titulo_alvo"`
	Resumo      string            `json:"resumo"`
	Skills      []string          `json:"skills"`
	Experiences []ExperienceInput `json:"experiences"`
	Education   []EducationInput  `json:"education"`
}

// ExperienceInput is one job entry from the editor.
type ExperienceInput struct {
	Empresa   string `json:"empresa"`
	Cargo     string `json:"cargo"`
	Inicio    string `json:"inicio"`
	Fim       string `json:"fim"`
	Descricao string `json:"descricao"`
}

// EducationInput is one study entry from the editor.
type EducationInput struct {
	Instituicao string `json:"instituicao"`
	Curso       string `json:"curso"`
	Inicio      string `json:"inicio"`
	Fim         string `json:"fim"`
}

// SaveCV validates and fully replaces the master CV.
func (s *Service) SaveCV(ctx context.Context, in CVInput) (*Profile, error) {
	p := &Profile{
		Nome:       strings.TrimSpace(in.Nome),
		Titulo:     strings.TrimSpace(in.Titulo),
		TituloAlvo: strings.TrimSpace(in.TituloAlvo),
		Resumo:     strings.TrimSpace(in.Resumo),
		Skills:     normalizeSkills(in.Skills),
	}
	for _, e := range in.Experiences {
		exp := Experience{
			Empresa:   strings.TrimSpace(e.Empresa),
			Cargo:     strings.TrimSpace(e.Cargo),
			Inicio:    strings.TrimSpace(e.Inicio),
			Fim:       strings.TrimSpace(e.Fim),
			Descricao: strings.TrimSpace(e.Descricao),
		}
		if exp.Empresa == "" && exp.Cargo == "" && exp.Descricao == "" {
			continue // skip blank rows
		}
		p.Experiences = append(p.Experiences, exp)
	}
	for _, e := range in.Education {
		ed := Education{
			Instituicao: strings.TrimSpace(e.Instituicao),
			Curso:       strings.TrimSpace(e.Curso),
			Inicio:      strings.TrimSpace(e.Inicio),
			Fim:         strings.TrimSpace(e.Fim),
		}
		if ed.Instituicao == "" && ed.Curso == "" {
			continue
		}
		p.Education = append(p.Education, ed)
	}
	if err := validateProfile(p); err != nil {
		return nil, err
	}
	if err := s.repo.SaveCV(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.GetProfile(ctx)
}

func validateProfile(p *Profile) error {
	for _, f := range []struct{ value, name string }{
		{p.Nome, "nome"}, {p.Titulo, "titulo"}, {p.TituloAlvo, "titulo_alvo"},
	} {
		if len([]rune(f.value)) > 120 {
			return fmt.Errorf("%w: %s muito longo (max 120)", ErrInvalid, f.name)
		}
	}
	if len([]rune(p.Resumo)) > 2000 {
		return fmt.Errorf("%w: resumo muito longo (max 2000)", ErrInvalid)
	}
	if len(p.Skills) > 100 {
		return fmt.Errorf("%w: muitas skills (max 100)", ErrInvalid)
	}
	if len(p.Experiences) > 30 {
		return fmt.Errorf("%w: muitas experiências (max 30)", ErrInvalid)
	}
	for _, e := range p.Experiences {
		if len([]rune(e.Empresa)) > 120 || len([]rune(e.Cargo)) > 120 {
			return fmt.Errorf("%w: empresa/cargo muito longo (max 120)", ErrInvalid)
		}
		if len([]rune(e.Descricao)) > 2000 {
			return fmt.Errorf("%w: descrição da experiência muito longa (max 2000)", ErrInvalid)
		}
	}
	if len(p.Education) > 20 {
		return fmt.Errorf("%w: muitas formações (max 20)", ErrInvalid)
	}
	for _, e := range p.Education {
		if len([]rune(e.Instituicao)) > 120 || len([]rune(e.Curso)) > 120 {
			return fmt.Errorf("%w: instituição/curso muito longo (max 120)", ErrInvalid)
		}
	}
	return nil
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
