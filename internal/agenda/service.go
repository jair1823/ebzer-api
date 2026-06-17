package agenda

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto CreateAgendaItemDTO) (int, error)
	GetByID(ctx context.Context, id int) (*AgendaItem, error)
	GetAll(ctx context.Context, filter FilterAgendaItemsDTO) ([]AgendaItem, error)
	Update(ctx context.Context, id int, dto UpdateAgendaItemDTO) error
	Delete(ctx context.Context, id int, actorID *int) error
	Complete(ctx context.Context, id int) error
	Archive(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// -------------------- Create --------------------

func (s *service) Create(ctx context.Context, dto CreateAgendaItemDTO) (int, error) {
	if dto.Title == "" {
		return 0, errors.New("title is required")
	}

	if dto.Type == "" {
		dto.Type = TypeNote
	} else if err := validateType(dto.Type); err != nil {
		return 0, err
	}

	if dto.Priority == "" {
		dto.Priority = PriorityMedium
	} else if err := validatePriority(dto.Priority); err != nil {
		return 0, err
	}

	if dto.Type == TypeReminder && dto.DueDate == nil {
		return 0, errors.New("due_date is required for reminder items")
	}

	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*AgendaItem, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll --------------------

func (s *service) GetAll(ctx context.Context, filter FilterAgendaItemsDTO) ([]AgendaItem, error) {
	// default status filter to "pending"
	if filter.Status == "" {
		filter.Status = string(StatusPending)
	}

	if filter.Status != "all" {
		if err := validateStatus(ItemStatus(filter.Status)); err != nil {
			return nil, err
		}
	}

	if filter.Type != "" {
		if err := validateType(ItemType(filter.Type)); err != nil {
			return nil, err
		}
	}

	if filter.Priority != "" {
		if err := validatePriority(ItemPriority(filter.Priority)); err != nil {
			return nil, err
		}
	}

	if filter.From != nil {
		if _, err := time.Parse("2006-01-02", *filter.From); err != nil {
			return nil, errors.New("invalid from date (expected format: YYYY-MM-DD)")
		}
	}

	if filter.To != nil {
		if _, err := time.Parse("2006-01-02", *filter.To); err != nil {
			return nil, errors.New("invalid to date (expected format: YYYY-MM-DD)")
		}
	}

	return s.repo.GetAll(ctx, filter)
}

// -------------------- Update --------------------

func (s *service) Update(ctx context.Context, id int, dto UpdateAgendaItemDTO) error {
	if dto.Title != nil && *dto.Title == "" {
		return errors.New("title cannot be empty")
	}

	if dto.Type != nil {
		if err := validateType(*dto.Type); err != nil {
			return err
		}
	}

	if dto.Status != nil {
		if err := validateStatus(*dto.Status); err != nil {
			return err
		}
	}

	if dto.Priority != nil {
		if err := validatePriority(*dto.Priority); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, id, dto)
}

// -------------------- Delete --------------------

func (s *service) Delete(ctx context.Context, id int, actorID *int) error {
	return s.repo.Delete(ctx, id, actorID)
}

// -------------------- Complete --------------------

func (s *service) Complete(ctx context.Context, id int) error {
	return s.repo.Complete(ctx, id)
}

// -------------------- Archive --------------------

func (s *service) Archive(ctx context.Context, id int) error {
	return s.repo.Archive(ctx, id)
}

// -------------------- Validation helpers --------------------

func validateType(t ItemType) error {
	switch t {
	case TypeNote, TypeTask, TypeReminder:
		return nil
	}
	return errors.New("type must be note, task, or reminder")
}

func validateStatus(s ItemStatus) error {
	switch s {
	case StatusPending, StatusDone, StatusArchived:
		return nil
	}
	return errors.New("status must be pending, done, or archived")
}

func validatePriority(p ItemPriority) error {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return nil
	}
	return errors.New("priority must be low, medium, or high")
}
