package intake

import (
	"context"

	idgen "github.com/lumni/mirante/internal/platform/id"
)

// SkillsProvider yields the owner's master CV skills for scoring. Injected by the
// composition root so intake stays decoupled from the cv domain (ADR-0001). It may
// be nil; scoring then treats skill fit as neutral.
type SkillsProvider func(ctx context.Context) ([]string, error)

// Service ingests opportunities into staging and exposes the triage funnel.
type Service struct {
	repo     Repository
	skills   SkillsProvider
	minScore int
}

// NewService builds the intake service. minScore is the shortlist floor used by
// callers that want only the strong leads; items below it are still staged.
func NewService(repo Repository, skills SkillsProvider, minScore int) *Service {
	return &Service{repo: repo, skills: skills, minScore: minScore}
}

// MinScore is the configured shortlist floor.
func (s *Service) MinScore() int { return s.minScore }

// Summary reports the outcome of an ingest run.
type Summary struct {
	Emails    int `json:"emails"`    // e-mails parsed successfully
	Failed    int `json:"failed"`    // e-mails that could not be parsed
	New       int `json:"new"`       // projects newly staged
	Duplicate int `json:"duplicate"` // projects already staged (deduped)
}

// Ingest parses raw feed e-mails, scores each project against the owner's skills,
// and stages the new ones (dedup by source id). It is resilient: an unparseable
// e-mail is counted and skipped, never fatal. Dedup keeps the first sighting, so a
// project recurring across daily digests does not reset its triage state.
func (s *Service) Ingest(ctx context.Context, rawEmails [][]byte) (Summary, error) {
	var mySkills []string
	if s.skills != nil {
		mySkills, _ = s.skills(ctx) // best-effort; Score tolerates an empty list
	}

	var sum Summary
	for _, raw := range rawEmails {
		projects, err := ParseEmail(raw)
		if err != nil {
			sum.Failed++
			continue
		}
		sum.Emails++
		for _, p := range projects {
			if p.FonteID == "" {
				continue // no dedup key → cannot stage safely
			}
			inserted, err := s.repo.Upsert(ctx, newItem(p, Score(p, mySkills)))
			if err != nil {
				return sum, err
			}
			if inserted {
				sum.New++
			} else {
				sum.Duplicate++
			}
		}
	}
	return sum, nil
}

// newItem builds a staged Item from a parsed project and its triage verdict. The
// stored skills are the project's detected skills (the owner's matches plus the
// gaps), for display chips.
func newItem(p ParsedProject, tri Triage) *Item {
	detected := append(append([]string{}, tri.Matched...), tri.Missing...)
	return &Item{
		ID:            ID(idgen.New()),
		Fonte:         p.Fonte,
		FonteID:       p.FonteID,
		Titulo:        p.Titulo,
		Categoria:     p.Categoria,
		Nivel:         p.Nivel,
		Publicado:     p.Publicado,
		TempoRestante: p.TempoRestante,
		RestanteHoras: p.RestanteHoras,
		Propostas:     p.Propostas,
		Interessados:  p.Interessados,
		Teaser:        p.Teaser,
		URL:           p.URL,
		EnviarURL:     p.EnviarURL,
		Skills:        detected,
		Score:         tri.Score,
		Estado:        EstadoNovo,
	}
}

// List returns staged items (filtered), highest score first.
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Item, error) {
	return s.repo.List(ctx, f)
}

// Get returns one staged item.
func (s *Service) Get(ctx context.Context, id ID) (*Item, error) {
	return s.repo.Get(ctx, id)
}

// Dismiss marks an item discarded so it drops out of the active funnel.
func (s *Service) Dismiss(ctx context.Context, id ID) error {
	return s.repo.SetEstado(ctx, id, EstadoDescartado)
}
