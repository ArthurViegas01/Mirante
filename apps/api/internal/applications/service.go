package applications

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

const dataLayout = "2006-01-02"

// Service holds candidatura use cases.
type Service struct{ repo Repository }

// NewService builds the applications service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// CreateInput is the payload for tracking a new candidatura.
type CreateInput struct {
	JobID       string `json:"job_id"`
	Titulo      string `json:"titulo"`
	Empresa     string `json:"empresa"`
	Status      Status `json:"status"`
	Notas       string `json:"notas"`
	ProximaAcao string `json:"proxima_acao"`
	DataAcao    string `json:"data_acao"`
}

// UpdateInput is a partial update; nil fields are left unchanged.
type UpdateInput struct {
	Titulo      *string `json:"titulo"`
	Empresa     *string `json:"empresa"`
	Status      *Status `json:"status"`
	Notas       *string `json:"notas"`
	ProximaAcao *string `json:"proxima_acao"`
	DataAcao    *string `json:"data_acao"`
}

// Get returns one candidatura.
func (s *Service) Get(ctx context.Context, id ID) (*Application, error) {
	return s.repo.Get(ctx, id)
}

// List returns candidaturas (optionally filtered by status).
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Application, error) {
	return s.repo.List(ctx, f)
}

// Create validates and persists a candidatura.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Application, error) {
	a := &Application{
		ID:          ID(idgen.New()),
		JobID:       strings.TrimSpace(in.JobID),
		Titulo:      strings.TrimSpace(in.Titulo),
		Empresa:     strings.TrimSpace(in.Empresa),
		Status:      in.Status,
		Notas:       in.Notas,
		ProximaAcao: strings.TrimSpace(in.ProximaAcao),
		DataAcao:    strings.TrimSpace(in.DataAcao),
	}
	if a.Status == "" {
		a.Status = StatusInteresse
	}
	if err := validateApplication(a); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, a.ID)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id ID, in UpdateInput) (*Application, error) {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Titulo != nil {
		a.Titulo = strings.TrimSpace(*in.Titulo)
	}
	if in.Empresa != nil {
		a.Empresa = strings.TrimSpace(*in.Empresa)
	}
	if in.Status != nil {
		a.Status = *in.Status
	}
	if in.Notas != nil {
		a.Notas = *in.Notas
	}
	if in.ProximaAcao != nil {
		a.ProximaAcao = strings.TrimSpace(*in.ProximaAcao)
	}
	if in.DataAcao != nil {
		a.DataAcao = strings.TrimSpace(*in.DataAcao)
	}
	if err := validateApplication(a); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

// Delete removes a candidatura.
func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

func validateApplication(a *Application) error {
	if n := strings.TrimSpace(a.Titulo); n == "" || len([]rune(n)) > 200 {
		return fmt.Errorf("%w: titulo é obrigatório (max 200)", ErrInvalid)
	}
	if err := validate.Var(string(a.Status), "oneof=interesse aplicado entrevista oferta aceito rejeitado"); err != nil {
		return fmt.Errorf("%w: status inválido", ErrInvalid)
	}
	if a.DataAcao != "" {
		if _, err := time.Parse(dataLayout, a.DataAcao); err != nil {
			return fmt.Errorf("%w: data_acao deve ser YYYY-MM-DD", ErrInvalid)
		}
	}
	if len([]rune(a.ProximaAcao)) > 200 {
		return fmt.Errorf("%w: proxima_acao muito longa (max 200)", ErrInvalid)
	}
	if len([]rune(a.Notas)) > 2000 {
		return fmt.Errorf("%w: notas muito longas (max 2000)", ErrInvalid)
	}
	return nil
}
