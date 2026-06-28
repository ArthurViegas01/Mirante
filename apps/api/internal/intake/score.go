package intake

import (
	"fmt"
	"strings"

	"github.com/lumni/mirante/internal/skills"
)

// Triage is the computed triage verdict for a parsed project: a 0..100 score plus
// the signals behind it, so the UI can both rank and explain.
type Triage struct {
	Score   int      `json:"score"`
	Matched []string `json:"matched"` // your skills the project asks for
	Missing []string `json:"missing"` // project skills you don't list
	Reasons []string `json:"reasons"` // human-readable signals
}

// Triage weights (skill fit, competition, freshness), summing to 1. Nível and
// categoria targeting are deliberately left out of this first cut — they need
// owner configuration the rest of the pipeline doesn't have yet.
const (
	wSkill = 0.45
	wComp  = 0.35
	wFresh = 0.20
)

// Score ranks a parsed project for the owner. The project's skills are detected
// from its title + teaser + the explicit "Habilidades desejadas" line via the
// shared skills kernel; aderência is their overlap with the owner's master
// skills. Competition (fewer propostas is better) and freshness (posted today
// beats older) complete the composite. With no detectable skills the skill term
// is neutral, so a project is never buried just for being hard to read.
func Score(p ParsedProject, mySkills []string) Triage {
	projectSkills := detectSkills(p)
	matched, missing := overlap(projectSkills, mySkills)

	skillScore := 50.0 // neutral when we can't read any skills
	if len(projectSkills) > 0 {
		skillScore = float64(len(matched)) / float64(len(projectSkills)) * 100
	}

	score := skillScore*wSkill + competitionScore(p.Propostas)*wComp + freshnessScore(p.Publicado)*wFresh
	return Triage{
		Score:   int(score + 0.5),
		Matched: matched,
		Missing: missing,
		Reasons: reasons(p, matched, projectSkills),
	}
}

// detectSkills canonicalizes everything the e-mail reveals about the project's
// stack: skills mentioned anywhere in the title/teaser, plus the explicit
// "Habilidades desejadas" entries (kept even when the catalog doesn't know them).
func detectSkills(p ParsedProject) []string {
	out := skills.Match(p.Titulo + "\n" + p.Teaser)
	seen := map[string]bool{}
	for _, s := range out {
		seen[strings.ToLower(s)] = true
	}
	for _, raw := range p.Skills {
		s := strings.TrimSpace(raw)
		if canon, ok := skills.Normalize(s); ok {
			s = canon
		}
		if s != "" && !seen[strings.ToLower(s)] {
			seen[strings.ToLower(s)] = true
			out = append(out, s)
		}
	}
	return out
}

// overlap splits projectSkills into those the owner lists and those they don't,
// comparing case-insensitively.
func overlap(projectSkills, mySkills []string) (matched, missing []string) {
	mine := map[string]bool{}
	for _, s := range mySkills {
		mine[strings.ToLower(strings.TrimSpace(s))] = true
	}
	matched, missing = []string{}, []string{}
	for _, s := range projectSkills {
		if mine[strings.ToLower(s)] {
			matched = append(matched, s)
		} else {
			missing = append(missing, s)
		}
	}
	return matched, missing
}

// competitionScore rewards low competition: 0 propostas → 100, 50+ propostas → 0.
func competitionScore(propostas int) float64 {
	if s := 100 - float64(propostas)*2; s > 0 {
		return s
	}
	return 0
}

// freshnessScore rewards recency from the relative "Publicado" text.
func freshnessScore(publicado string) float64 {
	switch {
	case strings.Contains(publicado, "hoje"):
		return 100
	case strings.Contains(publicado, "ontem"):
		return 60
	default:
		return 30
	}
}

// reasons builds the short human-readable signals shown alongside the score.
func reasons(p ParsedProject, matched, projectSkills []string) []string {
	var r []string
	switch {
	case p.Propostas <= 10:
		r = append(r, fmt.Sprintf("baixa concorrência (%d propostas)", p.Propostas))
	case p.Propostas >= 40:
		r = append(r, fmt.Sprintf("alta concorrência (%d propostas)", p.Propostas))
	}
	if strings.Contains(p.Publicado, "hoje") {
		r = append(r, "publicado hoje")
	}
	switch {
	case len(matched) > 0:
		r = append(r, "skills em comum: "+strings.Join(matched, ", "))
	case len(projectSkills) > 0:
		r = append(r, "nenhuma skill em comum")
	}
	return r
}
