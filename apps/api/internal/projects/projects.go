// Package projects is the central registry domain — the spine everything links
// to. It owns projects, their links, and the shared tag vocabulary.
package projects

import "time"

// ID is a project identifier.
type ID string

// Status is the lifecycle of a project.
type Status string

const (
	StatusIdeia     Status = "ideia"
	StatusAtivo     Status = "ativo"
	StatusPausado   Status = "pausado"
	StatusNoAr      Status = "no_ar"
	StatusArquivado Status = "arquivado"
)

// Visibility marks who a project is for.
type Visibility string

const (
	VisPessoal Visibility = "pessoal"
	VisLumni   Visibility = "lumni"
	VisCliente Visibility = "cliente"
)

// Project is the central entity.
type Project struct {
	ID           ID         `json:"id"`
	Nome         string     `json:"nome"`
	Codinome     string     `json:"codinome"`
	Descricao    string     `json:"descricao"`
	Repo         string     `json:"repo"`
	Status       Status     `json:"status"`
	Visibilidade Visibility `json:"visibilidade"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Links        []Link     `json:"links"`
	Tags         []string   `json:"tags"`
}

// Link is a typed external link attached to a project.
type Link struct {
	ID        ID        `json:"id"`
	ProjectID ID        `json:"project_id"`
	Label     string    `json:"label"`
	URL       string    `json:"url"`
	Kind      string    `json:"kind"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}
