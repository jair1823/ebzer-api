package orders

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto CreateOrderDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Order, error)
	GetAll(ctx context.Context, filter OrderFilterDTO) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error)
	Delete(ctx context.Context, id int, actorID *int) error
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
	if dto.Platform == "" {
		dto.Platform = PlatformWhatsApp
	}
	return s.repo.Create(ctx, dto)
}

// -------------------- GetByID --------------------

func (s *service) GetByID(ctx context.Context, id int) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GetAll with filters --------------------

func (s *service) GetAll(ctx context.Context, filterDTO OrderFilterDTO) ([]Order, error) {
	filter := OrderFilter{
		StatusID:      filterDTO.StatusID,
		StatusIDs:     filterDTO.StatusIDs,
		Search:        filterDTO.Search,
		Platform:      filterDTO.Platform,
		PaymentStatus: filterDTO.PaymentStatus,
		Overdue:       filterDTO.Overdue,
		AmountMin:     filterDTO.AmountMin,
		AmountMax:     filterDTO.AmountMax,
		Today:         time.Now(),
	}

	var err error
	if filter.From, err = parseDatePtr(filterDTO.From, "from"); err != nil {
		return nil, err
	}
	if filter.To, err = parseDatePtr(filterDTO.To, "to"); err != nil {
		return nil, err
	}
	if filter.DeliveryFrom, err = parseDatePtr(filterDTO.DeliveryFrom, "delivery_from"); err != nil {
		return nil, err
	}
	if filter.DeliveryTo, err = parseDatePtr(filterDTO.DeliveryTo, "delivery_to"); err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx, filter)
}

// -------------------- Update --------------------

func (s *service) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	return s.repo.Update(ctx, id, dto)
}

// -------------------- Delete --------------------

func (s *service) Delete(ctx context.Context, id int, actorID *int) error {
	return s.repo.Delete(ctx, id, actorID)
}

// -------------------- Finish Order --------------------

func (s *service) FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error) {
	return s.repo.FinishOrder(ctx, id)
}

func parseDatePtr(value *string, field string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, errors.New("invalid " + field + " date (expected format: YYYY-MM-DD)")
	}
	return &t, nil
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
