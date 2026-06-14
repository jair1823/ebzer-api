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
	GetAll(ctx context.Context, statuses []OrderStatus, from, to *string) ([]Order, error)
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
	if dto.ClientName == "" {
		return 0, errors.New("client_name is required")
	}

	// Aplicar defaults según schema
	if dto.Status == nil {
		newStatus := StatusNew
		dto.Status = &newStatus
	}

	if dto.DeliveryType == nil {
		pickup := DeliveryPickup
		dto.DeliveryType = &pickup
	}

	// Validar valores
	if !isValidOrderStatus(*dto.Status) {
		return 0, errors.New("invalid order status")
	}

	if !isValidDeliveryType(*dto.DeliveryType) {
		return 0, errors.New("invalid delivery type")
	}

	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll with filters --------------------

func (s *service) GetAll(ctx context.Context, statuses []OrderStatus, fromStr, toStr *string) ([]Order, error) {

	var from *time.Time
	var to *time.Time

	// Parse from (inicio del día)
	if fromStr != nil {
		t, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			return nil, errors.New("invalid from date (expected format: YYYY-MM-DD)")
		}
		from = &t
	}

	// Parse to (final del día - añadir 24 horas para incluir todo el día)
	if toStr != nil {
		t, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			return nil, errors.New("invalid to date (expected format: YYYY-MM-DD)")
		}
		// Añadir 1 día completo para incluir todo el día especificado
		// Ejemplo: "2026-04-22" -> "2026-04-22 00:00:00" + 24h = "2026-04-23 00:00:00"
		// Esto permite que entry_date <= "2026-04-23 00:00:00" incluya todo el día 22
		endOfDay := t.AddDate(0, 0, 1)
		to = &endOfDay
	}

	return s.repo.GetAll(ctx, statuses, from, to)
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

// -------------------- Validation Helpers --------------------

func isValidOrderStatus(status OrderStatus) bool {
	switch status {
	case StatusNew, StatusActive, StatusReady, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

func isValidDeliveryType(deliveryType DeliveryType) bool {
	switch deliveryType {
	case DeliveryPickup, DeliveryShipping, DeliveryDelivery:
		return true
	default:
		return false
	}
}
