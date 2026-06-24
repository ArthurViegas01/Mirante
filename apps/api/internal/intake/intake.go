// Package intake ingests freelance opportunity feeds (today: 99Freelas digest
// e-mails) into the career-search funnel. A single notification e-mail lists many
// projects; ParseDigest turns it into individual opportunities carrying the triage
// metadata — category, level, deadline, competition — that the scheduler scores
// before any are promoted to a tracked vaga. Per ADR-0001 it does not import other
// domains; it consumes the shared skills kernel and, later, the LLM gateway and
// the SSRF-guarded fetcher (for the full brief behind "Ver projeto").
//
// ParseDigest works on the e-mail's text rendering and extracts every field the
// triage needs. Pulling the per-project links (Ver projeto / Enviar proposta) out
// of the HTML part — needed for the authenticated full-brief fetch and for
// dedup-by-project-id — lands once a raw .eml sample pins down the markup.
package intake

import (
	"regexp"
	"strconv"
	"strings"
)

// ParsedProject is one freelance project extracted from a 99Freelas digest. The
// full brief (behind "Ver projeto") and the source links are filled in a later
// stage; everything here comes straight from the notification e-mail.
type ParsedProject struct {
	Titulo        string   `json:"titulo"`
	Categoria     string   `json:"categoria"`
	Nivel         string   `json:"nivel"`          // Iniciante | Intermediário | Especialista | ""
	Publicado     string   `json:"publicado"`      // raw, e.g. "hoje às 03:06"
	TempoRestante string   `json:"tempo_restante"` // raw, e.g. "5 dias e 19 horas"
	RestanteHoras int      `json:"restante_horas"` // best-effort parse of TempoRestante, in hours
	Propostas     int      `json:"propostas"`      // competition signal (lower = hotter lead)
	Interessados  int      `json:"interessados"`
	Teaser        string   `json:"teaser"` // truncated description, "Leia mais." trimmed
	Skills        []string `json:"skills"` // from "Habilidades desejadas:" when the e-mail lists it

	// Source + dedup metadata, filled by ParseEmail (empty from text-only ParseDigest).
	Fonte     string `json:"fonte"`      // e.g. "99freelas"
	FonteID   string `json:"fonte_id"`   // stable project id from the URL, the dedup key
	URL       string `json:"url"`        // "Ver projeto" link
	EnviarURL string `json:"enviar_url"` // "Enviar proposta" (bid) link
}

const (
	leiaMais   = "Leia mais."
	verProjeto = "Ver projeto"
	habil      = "Habilidades desejadas:"
)

var (
	// metaLineRe matches the one reliably-formatted line under each project:
	// "Categoria | Nível | Publicado: … | Tempo restante: … | Propostas: N | Interessados: M".
	// "Tempo restante:" before "Propostas:" is distinctive enough to anchor on.
	metaLineRe = regexp.MustCompile(`Tempo restante:.*Propostas:`)
	diasRe     = regexp.MustCompile(`(\d+)\s*dia`)
	horasRe    = regexp.MustCompile(`(\d+)\s*hora`)
	intRe      = regexp.MustCompile(`\d+`)
)

// ParseDigest extracts every project listed in a 99Freelas notification e-mail.
// It is lenient and best-effort: input is the e-mail's text rendering, each block
// is anchored on its metadata line (the only reliably-formatted line), and lines
// without one — the greeting, the section headers — are ignored. Projects are
// returned in the order they appear.
func ParseDigest(body string) []ParsedProject {
	var lines []string
	for _, l := range strings.Split(strings.ReplaceAll(body, "\r", ""), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			lines = append(lines, t)
		}
	}

	var out []ParsedProject
	for i, line := range lines {
		if i == 0 || !metaLineRe.MatchString(line) {
			continue // no metadata line, or no title above it
		}
		p := ParsedProject{Titulo: lines[i-1]}
		parseMeta(&p, line)

		// Teaser and skills sit on the lines after the metadata, up to the links.
		for _, next := range lines[i+1:] {
			if strings.HasPrefix(next, verProjeto) || metaLineRe.MatchString(next) {
				break
			}
			if strings.HasPrefix(next, habil) {
				p.Skills = splitSkills(strings.TrimPrefix(next, habil))
				continue
			}
			if p.Teaser == "" {
				p.Teaser = cleanTeaser(next)
			}
		}
		out = append(out, p)
	}
	return out
}

// parseMeta fills the structured fields from the "|"-separated metadata line.
// The first two segments are positional (categoria, nível); the rest are matched
// by their label so reordering or missing fields degrade gracefully.
func parseMeta(p *ParsedProject, line string) {
	for i, seg := range strings.Split(line, "|") {
		s := strings.TrimSpace(seg)
		switch {
		case strings.HasPrefix(s, "Publicado:"):
			p.Publicado = strings.TrimSpace(strings.TrimPrefix(s, "Publicado:"))
		case strings.HasPrefix(s, "Tempo restante:"):
			p.TempoRestante = strings.TrimSpace(strings.TrimPrefix(s, "Tempo restante:"))
			p.RestanteHoras = parseHoras(p.TempoRestante)
		case strings.HasPrefix(s, "Propostas:"):
			p.Propostas = firstInt(s)
		case strings.HasPrefix(s, "Interessados:"):
			p.Interessados = firstInt(s)
		case i == 0:
			p.Categoria = s
		case i == 1 && !strings.Contains(s, ":"):
			p.Nivel = s
		}
	}
}

// parseHoras turns "5 dias e 19 horas" (or "29 dias", "13 horas") into total hours.
func parseHoras(s string) int {
	h := 0
	if m := diasRe.FindStringSubmatch(s); m != nil {
		d, _ := strconv.Atoi(m[1])
		h += d * 24
	}
	if m := horasRe.FindStringSubmatch(s); m != nil {
		hh, _ := strconv.Atoi(m[1])
		h += hh
	}
	return h
}

// firstInt returns the first run of digits in s as an int (0 if none).
func firstInt(s string) int {
	n, _ := strconv.Atoi(intRe.FindString(s))
	return n
}

// cleanTeaser trims the trailing "Leia mais." call-to-action, keeping the "…"
// truncation marker that signals the brief continues behind the link.
func cleanTeaser(s string) string {
	if i := strings.LastIndex(s, leiaMais); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// splitSkills splits a comma-separated "Habilidades desejadas" list, trimming blanks.
func splitSkills(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}
