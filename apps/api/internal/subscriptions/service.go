package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/validate"
)

// ErrInvalid wraps validation failures (mapped to HTTP 400).
var ErrInvalid = errors.New("invalid input")

// Service holds subscription use cases.
type Service struct {
	repo Repository
}

// NewService builds the subscriptions service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// CreateInput is the payload for creating a subscription.
type CreateInput struct {
	ProjectID  string   `json:"project_id"`
	ServiceID  string   `json:"service_id"`
	Nome       string   `json:"nome"`
	Provider   string   `json:"provider"`
	ValorCents int      `json:"valor_cents"`
	Moeda      Currency `json:"moeda"`
	Ciclo      Cycle    `json:"ciclo"`
	Ativo      *bool    `json:"ativo"`
	Notas      string   `json:"notas"`
}

// UpdateInput is a partial update; nil fields are left unchanged.
type UpdateInput struct {
	ServiceID  *string   `json:"service_id"`
	Nome       *string   `json:"nome"`
	Provider   *string   `json:"provider"`
	ValorCents *int      `json:"valor_cents"`
	Moeda      *Currency `json:"moeda"`
	Ciclo      *Cycle    `json:"ciclo"`
	Ativo      *bool     `json:"ativo"`
	Notas      *string   `json:"notas"`
}

// Get returns a subscription.
func (s *Service) Get(ctx context.Context, id ID) (*Subscription, error) {
	return s.repo.Get(ctx, id)
}

// List returns subscriptions (optionally scoped to a project).
func (s *Service) List(ctx context.Context, f ListFilter) ([]*Subscription, error) {
	return s.repo.List(ctx, f)
}

// Create validates and persists a new subscription.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Subscription, error) {
	sub := &Subscription{
		ID:         ID(idgen.New()),
		ProjectID:  strings.TrimSpace(in.ProjectID),
		ServiceID:  strings.TrimSpace(in.ServiceID),
		Nome:       strings.TrimSpace(in.Nome),
		Provider:   strings.TrimSpace(in.Provider),
		ValorCents: in.ValorCents,
		Moeda:      in.Moeda,
		Ciclo:      in.Ciclo,
		Ativo:      in.Ativo == nil || *in.Ativo,
		Notas:      in.Notas,
	}
	if sub.Moeda == "" {
		sub.Moeda = MoedaBRL
	}
	if sub.Ciclo == "" {
		sub.Ciclo = CicloMensal
	}
	if err := validateSubscription(sub); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, sub.ID)
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id ID, in UpdateInput) (*Subscription, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.ServiceID != nil {
		sub.ServiceID = strings.TrimSpace(*in.ServiceID)
	}
	if in.Nome != nil {
		sub.Nome = strings.TrimSpace(*in.Nome)
	}
	if in.Provider != nil {
		sub.Provider = strings.TrimSpace(*in.Provider)
	}
	if in.ValorCents != nil {
		sub.ValorCents = *in.ValorCents
	}
	if in.Moeda != nil {
		sub.Moeda = *in.Moeda
	}
	if in.Ciclo != nil {
		sub.Ciclo = *in.Ciclo
	}
	if in.Ativo != nil {
		sub.Ativo = *in.Ativo
	}
	if in.Notas != nil {
		sub.Notas = *in.Notas
	}
	if err := validateSubscription(sub); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

// Delete removes a subscription.
func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

func validateSubscription(s *Subscription) error {
	if s.ProjectID == "" {
		return fmt.Errorf("%w: project_id é obrigatório", ErrInvalid)
	}
	if n := strings.TrimSpace(s.Nome); n == "" || len([]rune(n)) > 120 {
		return fmt.Errorf("%w: nome é obrigatório (max 120)", ErrInvalid)
	}
	if len([]rune(s.Provider)) > 40 {
		return fmt.Errorf("%w: provider muito longo (max 40)", ErrInvalid)
	}
	if s.ValorCents < 0 {
		return fmt.Errorf("%w: valor_cents não pode ser negativo", ErrInvalid)
	}
	if err := validate.Var(string(s.Moeda), "oneof=BRL USD"); err != nil {
		return fmt.Errorf("%w: moeda deve ser BRL ou USD", ErrInvalid)
	}
	if err := validate.Var(string(s.Ciclo), "oneof=mensal anual"); err != nil {
		return fmt.Errorf("%w: ciclo deve ser mensal ou anual", ErrInvalid)
	}
	return nil
}
