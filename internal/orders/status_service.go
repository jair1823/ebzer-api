package orders

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	slugRegex  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

// StatusService manages configurable order statuses
type StatusService interface {
	Create(ctx context.Context, dto CreateOrderStatusDTO) (int, error)
	GetByID(ctx context.Context, id int) (*OrderStatus, error)
	GetAll(ctx context.Context, activeOnly bool) ([]OrderStatus, error)
	Update(ctx context.Context, id int, dto UpdateOrderStatusDTO) error
	Deactivate(ctx context.Context, id int) error
	Reorder(ctx context.Context, dto ReorderStatusesDTO) error
}

type statusService struct {
	repo StatusRepository
}

func NewStatusService(repo StatusRepository) StatusService {
	return &statusService{repo: repo}
}

// -------------------- CREATE --------------------

func (s *statusService) Create(ctx context.Context, dto CreateOrderStatusDTO) (int, error) {
	dto.Name = strings.ToLower(strings.TrimSpace(dto.Name))

	if dto.Name == "" {
		return 0, errors.New("name is required")
	}
	if !slugRegex.MatchString(dto.Name) {
		return 0, errors.New("name must be lowercase letters, digits, and underscores (no leading digit)")
	}
	if strings.TrimSpace(dto.DisplayName) == "" {
		return 0, errors.New("display_name is required")
	}
	if dto.Color != "" && !colorRegex.MatchString(dto.Color) {
		return 0, errors.New("color must be a valid hex color (e.g. #3B82F6)")
	}
	if dto.OrderPosition <= 0 {
		return 0, errors.New("order_position must be > 0")
	}

	existing, err := s.repo.GetByName(ctx, dto.Name)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, fmt.Errorf("status with name %q already exists", dto.Name)
	}

	return s.repo.Create(ctx, dto)
}

// -------------------- GET BY ID --------------------

func (s *statusService) GetByID(ctx context.Context, id int) (*OrderStatus, error) {
	return s.repo.GetByID(ctx, id)
}

// -------------------- GET ALL --------------------

func (s *statusService) GetAll(ctx context.Context, activeOnly bool) ([]OrderStatus, error) {
	return s.repo.GetAll(ctx, activeOnly)
}

// -------------------- UPDATE --------------------

func (s *statusService) Update(ctx context.Context, id int, dto UpdateOrderStatusDTO) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("order status not found")
	}

	if dto.Color != nil && *dto.Color != "" && !colorRegex.MatchString(*dto.Color) {
		return errors.New("color must be a valid hex color (e.g. #3B82F6)")
	}
	if dto.OrderPosition != nil && *dto.OrderPosition <= 0 {
		return errors.New("order_position must be > 0")
	}

	// System statuses cannot be deactivated
	if existing.IsSystemStatus && dto.IsActive != nil && !*dto.IsActive {
		return fmt.Errorf("cannot deactivate system status %q", existing.Name)
	}

	return s.repo.Update(ctx, id, dto)
}

// -------------------- DEACTIVATE --------------------

func (s *statusService) Deactivate(ctx context.Context, id int) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("order status not found")
	}
	if existing.IsSystemStatus {
		return fmt.Errorf("cannot deactivate system status %q", existing.Name)
	}

	count, err := s.repo.CountOrdersForStatus(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot deactivate: %d order(s) currently use this status", count)
	}

	return s.repo.Deactivate(ctx, id)
}

// -------------------- REORDER --------------------

func (s *statusService) Reorder(ctx context.Context, dto ReorderStatusesDTO) error {
	if len(dto.StatusOrders) == 0 {
		return errors.New("status_orders is required and cannot be empty")
	}

	// Verify all positions are unique and positive
	seen := map[int]bool{}
	for _, item := range dto.StatusOrders {
		if item.Position <= 0 {
			return fmt.Errorf("position for id %d must be > 0", item.ID)
		}
		if seen[item.Position] {
			return fmt.Errorf("duplicate position %d", item.Position)
		}
		seen[item.Position] = true
	}

	return s.repo.Reorder(ctx, dto.StatusOrders)
}
