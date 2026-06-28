package intake

import "time"

// ID identifies a staged intake item.
type ID string

// Estado is the funnel lifecycle of a staged item. Score and "shortlist" are
// computed (score >= the configured floor), not stored as states — only the
// owner-driven lifecycle lives here.
type Estado string

const (
	EstadoNovo       Estado = "novo"       // ingested, awaiting triage
	EstadoDescartado Estado = "descartado" // dismissed by the owner
	EstadoPromovido  Estado = "promovido"  // turned into a tracked vaga + candidatura
)

// Item is a freelance opportunity staged from a feed and scored for triage. Most
// items are discarded; the chosen few are promoted out of staging. The dedup key
// is (user, Fonte, FonteID) — the same project recurs across daily digests.
type Item struct {
	ID            ID        `json:"id"`
	Fonte         string    `json:"fonte"`
	FonteID       string    `json:"fonte_id"`
	Titulo        string    `json:"titulo"`
	Categoria     string    `json:"categoria"`
	Nivel         string    `json:"nivel"`
	Publicado     string    `json:"publicado"`
	TempoRestante string    `json:"tempo_restante"`
	RestanteHoras int       `json:"restante_horas"`
	Propostas     int       `json:"propostas"`
	Interessados  int       `json:"interessados"`
	Teaser        string    `json:"teaser"`
	URL           string    `json:"url"`
	EnviarURL     string    `json:"enviar_url"`
	Skills        []string  `json:"skills"` // skills detected on the project
	Score         int       `json:"score"`  // triage score snapshot at ingest
	Estado        Estado    `json:"estado"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
