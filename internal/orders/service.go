package orders

import (
	"context"
	"errors"
	"time"

	"creaciones-api/internal/incomes"
)

type Service interface {
	Create(ctx context.Context, dto CreateOrderDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Order, error)
	GetAll(ctx context.Context, status *OrderStatus, from, to *string) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	FinishOrder(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
	GetPaymentStatus(ctx context.Context, orderID int) (*PaymentStatus, error)
}

type service struct {
	repo       Repository
	incomeRepo incomes.Repository
}

func NewService(repo Repository, incomeRepo incomes.Repository) Service {
	return &service{
		repo:       repo,
		incomeRepo: incomeRepo,
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
	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll with filters --------------------

func (s *service) GetAll(ctx context.Context, status *OrderStatus, fromStr, toStr *string) ([]Order, error) {

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

	return s.repo.GetAll(ctx, status, from, to)
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

func (s *service) FinishOrder(ctx context.Context, id int) error {
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

	// Get all incomes for this order
	orderIncomes, err := s.incomeRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Calculate total paid
	totalPaid := 0.0
	for _, income := range orderIncomes {
		totalPaid += income.Amount
	}

	// Calculate remaining and percentage
	remaining := order.AmountCharged - totalPaid
	percentagePaid := 0.0
	if order.AmountCharged > 0 {
		percentagePaid = (totalPaid / order.AmountCharged) * 100
	}

	return &PaymentStatus{
		TotalPaid:      totalPaid,
		AmountCharged:  order.AmountCharged,
		Remaining:      remaining,
		PercentagePaid: percentagePaid,
		IsFullyPaid:    totalPaid >= order.AmountCharged,
	}, nil
}
