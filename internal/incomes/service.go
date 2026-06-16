package incomes

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto CreateIncomeDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Income, error)
	GetAll(ctx context.Context, from, to *string) ([]Income, error)
	Update(ctx context.Context, id int, dto UpdateIncomeDTO) error
	Delete(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// -------------------- Create --------------------

func (s *service) Create(ctx context.Context, dto CreateIncomeDTO) (int, error) {
	if dto.OrderID <= 0 {
		return 0, errors.New("order ID is required and must be > 0")
	}
	if dto.Amount <= 0 {
		return 0, errors.New("amount must be > 0")
	}
	date, err := normalizeIncomeDate(dto.Date)
	if err != nil {
		return 0, err
	}
	dto.Date = date
	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*Income, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll with filters --------------------

func (s *service) GetAll(ctx context.Context, fromStr, toStr *string) ([]Income, error) {

	var from *time.Time
	var to *time.Time

	// Parse from
	if fromStr != nil {
		t, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			return nil, errors.New("invalid from date (expected format: YYYY-MM-DD)")
		}
		from = &t
	}

	// Parse to
	if toStr != nil {
		t, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			return nil, errors.New("invalid to date (expected format: YYYY-MM-DD)")
		}
		to = &t
	}

	return s.repo.GetAll(ctx, from, to)
}

// -------------------- Update --------------------

func (s *service) Update(ctx context.Context, id int, dto UpdateIncomeDTO) error {
	if id <= 0 {
		return errors.New("income ID is required and must be > 0")
	}
	if dto.OrderID != nil && *dto.OrderID <= 0 {
		return errors.New("order ID must be > 0")
	}
	if dto.Amount != nil && *dto.Amount <= 0 {
		return errors.New("amount must be > 0")
	}
	date, err := normalizeIncomeDate(dto.Date)
	if err != nil {
		return err
	}
	dto.Date = date
	return s.repo.Update(ctx, id, dto)
}

// -------------------- Delete --------------------

func (s *service) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("income ID is required and must be > 0")
	}
	return s.repo.Delete(ctx, id)
}

func normalizeIncomeDate(date *string) (*string, error) {
	if date == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*date)
	if trimmed == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return nil, errors.New("invalid date (expected format: YYYY-MM-DD)")
	}

	return &trimmed, nil
}
