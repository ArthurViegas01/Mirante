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

// prazoLayout is the calendar-date format accepted for a task deadline.
const prazoLayout = "2006-01-02"

// Service holds task use cases.
type Service struct {
	repo Repository
}

// NewService builds the tasks service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// CreateInput is the payload for creating a task.
type CreateInput struct {
	Titulo     string   `json:"titulo"`
	Descricao  string   `json:"descricao"`
	Status     Status   `json:"status"`
	Prioridade Priority `json:"prioridade"`
	Prazo      string   `json:"prazo"`
	ProjectID  string   `json:"project_id"`
	JobID      string   `json:"job_id"`
	Tags       []string `json:"tags"`
}

// UpdateInput is a partial update; nil fields are left unchanged.
type UpdateInput struct {
	Titulo     *string   `json:"titulo"`
	Descricao  *string   `json:"descricao"`
	Status     *Status   `json:"status"`
	Prioridade *Priority `json:"prioridade"`
	Prazo      *string   `json:"prazo"`
	ProjectID  *string   `json:"project_id"`
	JobID      *string   `json:"job_id"`
	Tags       *[]string `json:"tags"`
}

// Get returns a task with its tags.
func (s *Service) Get(ctx context.Context, id ID) (*Task, error) {
	return s.repo.Get(ctx, id)
}

// List returns tasks (optionally filtered by status and/or project).
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Task, error) {
	return s.repo.List(ctx, f)
}

// Create validates and persists a new task.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Task, error) {
	t := &Task{
		ID:         ID(idgen.New()),
		Titulo:     strings.TrimSpace(in.Titulo),
		Descricao:  in.Descricao,
		Status:     in.Status,
		Prioridade: in.Prioridade,
		Prazo:      strings.TrimSpace(in.Prazo),
		ProjectID:  strings.TrimSpace(in.ProjectID),
		JobID:      strings.TrimSpace(in.JobID),
	}
	if t.Status == "" {
		t.Status = StatusAFazer
	}
	if t.Prioridade == "" {
		t.Prioridade = PrioridadeMedia
	}
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
	if in.Descricao != nil {
		t.Descricao = *in.Descricao
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	if in.Prioridade != nil {
		t.Prioridade = *in.Prioridade
	}
	if in.Prazo != nil {
		t.Prazo = strings.TrimSpace(*in.Prazo)
	}
	if in.ProjectID != nil {
		t.ProjectID = strings.TrimSpace(*in.ProjectID)
	}
	if in.JobID != nil {
		t.JobID = strings.TrimSpace(*in.JobID)
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

// Delete hard-deletes a task (cascading its tags). A task is its own record;
// deleting a project only unlinks its tasks (project_id ON DELETE SET NULL).
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
	if t.Prazo != "" {
		if _, err := time.Parse(prazoLayout, t.Prazo); err != nil {
			return fmt.Errorf("%w: prazo must be YYYY-MM-DD", ErrInvalid)
		}
	}
	return nil
}
