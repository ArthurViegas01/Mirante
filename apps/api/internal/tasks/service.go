package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/validate"
)

// ErrInvalid wraps validation failures (mapped to HTTP 400).
var ErrInvalid = errors.New("invalid input")

// prazoLayout is the calendar-date layout for the nullable deadline.
const prazoLayout = "2006-01-02"

// Service holds task use cases.
type Service struct {
	repo Repository
}

// NewService builds the tasks service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// CreateInput is the payload for creating a task.
type CreateInput struct {
	Titulo     string     `json:"titulo"`
	Status     Status     `json:"status"`
	Prioridade Prioridade `json:"prioridade"`
	Prazo      *string    `json:"prazo"`
	ProjectID  *string    `json:"project_id"`
	JobID      *string    `json:"job_id"`
	Tags       []string   `json:"tags"`
}

// UpdateInput is a partial update. A nil field is left unchanged; a non-nil
// Prazo/ProjectID/JobID set to "" clears it to NULL.
type UpdateInput struct {
	Titulo     *string     `json:"titulo"`
	Status     *Status     `json:"status"`
	Prioridade *Prioridade `json:"prioridade"`
	Prazo      *string     `json:"prazo"`
	ProjectID  *string     `json:"project_id"`
	JobID      *string     `json:"job_id"`
	Tags       *[]string   `json:"tags"`
}

// Get returns a task with its tags.
func (s *Service) Get(ctx context.Context, id ID) (*Task, error) {
	return s.repo.Get(ctx, id)
}

// List returns tasks (optionally filtered by status, project and priority).
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Task, error) {
	return s.repo.List(ctx, f)
}

// Create validates and persists a new task.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Task, error) {
	t := &Task{
		ID:         ID(idgen.New()),
		Titulo:     strings.TrimSpace(in.Titulo),
		Status:     in.Status,
		Prioridade: in.Prioridade,
		ProjectID:  cleanPtr(in.ProjectID),
		JobID:      cleanPtr(in.JobID),
	}
	if t.Status == "" {
		t.Status = StatusAFazer
	}
	if t.Prioridade == "" {
		t.Prioridade = PrioMedia
	}
	prazo, err := normalizePrazo(in.Prazo)
	if err != nil {
		return nil, err
	}
	t.Prazo = prazo
	if err := validateTask(t); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	if len(in.Tags) > 0 {
		if err := s.repo.SetTags(ctx, t.ID, in.Tags); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, t.ID)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id ID, in UpdateInput) (*Task, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Titulo != nil {
		t.Titulo = strings.TrimSpace(*in.Titulo)
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	if in.Prioridade != nil {
		t.Prioridade = *in.Prioridade
	}
	if in.Prazo != nil {
		prazo, err := normalizePrazo(in.Prazo)
		if err != nil {
			return nil, err
		}
		t.Prazo = prazo
	}
	if in.ProjectID != nil {
		t.ProjectID = cleanPtr(in.ProjectID)
	}
	if in.JobID != nil {
		t.JobID = cleanPtr(in.JobID)
	}
	if err := validateTask(t); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	if in.Tags != nil {
		if err := s.repo.SetTags(ctx, id, *in.Tags); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, id)
}

// Delete hard-deletes a task (cascading its tag links).
func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

func validateTask(t *Task) error {
	if n := strings.TrimSpace(t.Titulo); n == "" || len(n) > 200 {
		return fmt.Errorf("%w: titulo is required (max 200)", ErrInvalid)
	}
	if err := validate.Var(string(t.Status), "oneof=a_fazer fazendo feito"); err != nil {
		return fmt.Errorf("%w: status", ErrInvalid)
	}
	if err := validate.Var(string(t.Prioridade), "oneof=baixa media alta"); err != nil {
		return fmt.Errorf("%w: prioridade", ErrInvalid)
	}
	return nil
}

// normalizePrazo trims, validates the YYYY-MM-DD form, and maps empty to nil.
func normalizePrazo(p *string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil, nil
	}
	if _, err := time.Parse(prazoLayout, s); err != nil {
		return nil, fmt.Errorf("%w: prazo deve ser uma data YYYY-MM-DD", ErrInvalid)
	}
	return &s, nil
}

// cleanPtr trims an optional string and maps empty to nil.
func cleanPtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return &s
}
