package orders

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto CreateOrderDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Order, error)
	GetAll(ctx context.Context, statusID *int, from, to *string) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error)
	Delete(ctx context.Context, id int) error
	GetPaymentStatus(ctx context.Context, orderID int) (*PaymentStatus, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// -------------------- Create --------------------

func (s *service) Create(ctx context.Context, dto CreateOrderDTO) (int, error) {
	if dto.Description == "" {
		return 0, errors.New("description is required")
	}
	if dto.AmountCharged < 0 {
		return 0, errors.New("amount_charged must be >= 0")
	}
	// Default to 'new' system status (id = 1 from seed) if not provided
	if dto.StatusID == 0 {
		dto.StatusID = 1
	}
	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll with filters --------------------

func (s *service) GetAll(ctx context.Context, statusID *int, fromStr, toStr *string) ([]Order, error) {

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

	return s.repo.GetAll(ctx, statusID, from, to)
}

// -------------------- Update --------------------

func (s *service) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	return s.repo.Update(ctx, id, dto)
}

// -------------------- Delete --------------------

func (s *service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// -------------------- Finish Order --------------------

func (s *service) FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error) {
	return s.repo.FinishOrder(ctx, id)
}

// -------------------- Get Payment Status --------------------

func (s *service) GetPaymentStatus(ctx context.Context, orderID int) (*PaymentStatus, error) {
	// Get order to validate it exists and get amount charged
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	return order.PaymentStatus, nil
}
