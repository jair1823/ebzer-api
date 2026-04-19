package orders

import (
	"context"
	"testing"
	"time"
)

// Mock repository para testing
type mockRepository struct {
	createFunc func(ctx context.Context, dto CreateOrderDTO) (int, error)
}

func (m *mockRepository) Create(ctx context.Context, dto CreateOrderDTO) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, dto)
	}
	return 1, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int) (*Order, error) {
	return nil, nil
}

func (m *mockRepository) GetAll(ctx context.Context, status *OrderStatus, from, to *time.Time) ([]Order, error) {
	return nil, nil
}

func (m *mockRepository) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	return nil
}

func (m *mockRepository) FinishOrder(ctx context.Context, id int) error {
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func TestCreate_AppliesDefaultStatus(t *testing.T) {
	var capturedDTO CreateOrderDTO

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, dto CreateOrderDTO) (int, error) {
			capturedDTO = dto
			return 1, nil
		},
	}

	service := NewService(mockRepo, nil)

	dto := CreateOrderDTO{
		Description:   "Test order",
		AmountCharged: 100.50,
		ClientName:    "Test Client",
		Status:        nil, // Omitido, debe usar default
		DeliveryType:  nil, // Omitido, debe usar default
	}

	_, err := service.Create(context.Background(), dto)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if capturedDTO.Status == nil {
		t.Fatal("Expected Status to be set with default value")
	}

	if *capturedDTO.Status != StatusConfirmed {
		t.Errorf("Expected default Status 'confirmed', got: %s", *capturedDTO.Status)
	}

	if capturedDTO.DeliveryType == nil {
		t.Fatal("Expected DeliveryType to be set with default value")
	}

	if *capturedDTO.DeliveryType != DeliveryPickup {
		t.Errorf("Expected default DeliveryType 'pickup', got: %s", *capturedDTO.DeliveryType)
	}
}

func TestCreate_RespectsProvidedStatus(t *testing.T) {
	var capturedDTO CreateOrderDTO

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, dto CreateOrderDTO) (int, error) {
			capturedDTO = dto
			return 1, nil
		},
	}

	service := NewService(mockRepo, nil)

	customStatus := StatusInProgress
	customDelivery := DeliveryShipping

	dto := CreateOrderDTO{
		Description:   "Test order",
		AmountCharged: 100.50,
		ClientName:    "Test Client",
		Status:        &customStatus,
		DeliveryType:  &customDelivery,
	}

	_, err := service.Create(context.Background(), dto)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if *capturedDTO.Status != StatusInProgress {
		t.Errorf("Expected Status 'in_progress', got: %s", *capturedDTO.Status)
	}

	if *capturedDTO.DeliveryType != DeliveryShipping {
		t.Errorf("Expected DeliveryType 'shipping', got: %s", *capturedDTO.DeliveryType)
	}
}

func TestCreate_RejectsInvalidStatus(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo, nil)

	invalidStatus := OrderStatus("invalid_status")

	dto := CreateOrderDTO{
		Description:   "Test order",
		AmountCharged: 100.50,
		ClientName:    "Test Client",
		Status:        &invalidStatus,
	}

	_, err := service.Create(context.Background(), dto)

	if err == nil {
		t.Fatal("Expected error for invalid status, got nil")
	}

	if err.Error() != "invalid order status" {
		t.Errorf("Expected 'invalid order status' error, got: %v", err)
	}
}

func TestCreate_RejectsInvalidDeliveryType(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo, nil)

	invalidDelivery := DeliveryType("invalid_delivery")

	dto := CreateOrderDTO{
		Description:   "Test order",
		AmountCharged: 100.50,
		ClientName:    "Test Client",
		DeliveryType:  &invalidDelivery,
	}

	_, err := service.Create(context.Background(), dto)

	if err == nil {
		t.Fatal("Expected error for invalid delivery type, got nil")
	}

	if err.Error() != "invalid delivery type" {
		t.Errorf("Expected 'invalid delivery type' error, got: %v", err)
	}
}
