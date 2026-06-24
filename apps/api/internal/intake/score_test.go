package intake

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScoreCompetitionAndFreshness(t *testing.T) {
	p := ParseDigest(digestFixture)
	require.Len(t, p, 11)

	// With no skills to weigh, the hot lead (posted today, 3 proposals) must
	// outrank the saturated one (74 proposals, posted yesterday).
	oficina := Score(p[2], nil) // oficina mecânica, 74 propostas, ontem
	hot := Score(p[10], nil)    // notificações de risco, 3 propostas, hoje
	require.Greater(t, hot.Score, oficina.Score)

	require.Contains(t, hot.Reasons, "publicado hoje")
	require.Contains(t, hot.Reasons, "baixa concorrência (3 propostas)")
	require.Contains(t, oficina.Reasons, "alta concorrência (74 propostas)")
}

func TestScoreSkillMatch(t *testing.T) {
	p := ParseDigest(digestFixture)
	mine := []string{"Shopify", "WordPress", "Go"}

	// Detected from the title/teaser text.
	require.Contains(t, Score(p[0], mine).Matched, "Shopify")
	// Detected from the explicit "Habilidades desejadas: Wordpress" line.
	require.Contains(t, Score(p[9], mine).Matched, "WordPress")

	// A required skill the owner lacks lands in Missing, not Matched.
	pentaho := Score(p[4], mine)
	require.Contains(t, pentaho.Missing, "Pentaho")
	require.NotContains(t, pentaho.Matched, "Pentaho")
}

func TestScoreRankingBuriesSaturated(t *testing.T) {
	p := ParseDigest(digestFixture)
	mine := []string{"Shopify", "WordPress"}

	type scored struct {
		titulo string
		score  int
	}
	ranked := make([]scored, len(p))
	for i, proj := range p {
		ranked[i] = scored{proj.Titulo, Score(proj, mine).Score}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	// The 74-proposal oficina job should fall into the bottom half of the ranking.
	oficinaRank := -1
	for i, r := range ranked {
		if r.titulo == "Sistema de gestão para oficina mecânica com orçamento via WhatsApp" {
			oficinaRank = i
		}
	}
	require.Greater(t, oficinaRank, 5, "saturated job should rank in the bottom half")
}
