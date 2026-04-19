package incomes

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock repository para testing
type mockRepository struct {
	createFunc     func(ctx context.Context, dto CreateIncomeDTO) (int, error)
	getByIDFunc    func(ctx context.Context, id int) (*Income, error)
	getAllFunc     func(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error)
	getByOrderFunc func(ctx context.Context, orderID int) ([]Income, error)
	updateFunc     func(ctx context.Context, id int, dto UpdateIncomeDTO) error
	deleteFunc     func(ctx context.Context, id int) error
}

func (m *mockRepository) Create(ctx context.Context, dto CreateIncomeDTO) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, dto)
	}
	return 1, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int) (*Income, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetAll(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx, from, to)
	}
	return []Income{}, nil
}

func (m *mockRepository) GetByOrderID(ctx context.Context, orderID int) ([]Income, error) {
	if m.getByOrderFunc != nil {
		return m.getByOrderFunc(ctx, orderID)
	}
	return []Income{}, nil
}

func (m *mockRepository) Update(ctx context.Context, id int, dto UpdateIncomeDTO) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, dto)
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// ========== Tests ==========

func TestCreate_ValidOrderID(t *testing.T) {
	var capturedDTO CreateIncomeDTO

	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, dto CreateIncomeDTO) (int, error) {
			capturedDTO = dto
			return 456, nil
		},
	}

	service := NewService(mockRepo)

	dto := CreateIncomeDTO{
		OrderID: 123,
		Amount:  CustomFloat64(500.00),
	}

	id, err := service.Create(context.Background(), dto)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if id != 456 {
		t.Errorf("Expected ID 456, got: %d", id)
	}

	if capturedDTO.OrderID != 123 {
		t.Errorf("Expected OrderID 123, got: %d", capturedDTO.OrderID)
	}

	if float64(capturedDTO.Amount) != 500.00 {
		t.Errorf("Expected Amount 500.00, got: %f", capturedDTO.Amount)
	}
}

func TestCreate_OrderIDZero_ShouldFail(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	dto := CreateIncomeDTO{
		OrderID: 0,
		Amount:  CustomFloat64(500.00),
	}

	_, err := service.Create(context.Background(), dto)

	if err == nil {
		t.Fatal("Expected error for OrderID = 0, got nil")
	}

	expectedMsg := "order ID is required and must be positive"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestCreate_OrderIDNegative_ShouldFail(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	dto := CreateIncomeDTO{
		OrderID: -5,
		Amount:  CustomFloat64(500.00),
	}

	_, err := service.Create(context.Background(), dto)

	if err == nil {
		t.Fatal("Expected error for negative OrderID, got nil")
	}

	expectedMsg := "order ID is required and must be positive"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestCreate_NegativeAmount_ShouldFail(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	dto := CreateIncomeDTO{
		OrderID: 123,
		Amount:  CustomFloat64(-100.00),
	}

	_, err := service.Create(context.Background(), dto)

	if err == nil {
		t.Fatal("Expected error for negative amount, got nil")
	}

	expectedMsg := "amount must be >= 0"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestCreate_ZeroAmount_ShouldSucceed(t *testing.T) {
	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, dto CreateIncomeDTO) (int, error) {
			return 1, nil
		},
	}
	service := NewService(mockRepo)

	dto := CreateIncomeDTO{
		OrderID: 123,
		Amount:  CustomFloat64(0),
	}

	_, err := service.Create(context.Background(), dto)

	if err != nil {
		t.Fatalf("Expected no error for amount = 0, got: %v", err)
	}
}

func TestGetAll_InvalidFromDate_ShouldFail(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	invalidDate := "2026-13-01" // Mes inválido

	_, err := service.GetAll(context.Background(), &invalidDate, nil)

	if err == nil {
		t.Fatal("Expected error for invalid 'from' date, got nil")
	}

	expectedMsg := "invalid from date (expected format: YYYY-MM-DD)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestGetAll_InvalidToDate_ShouldFail(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	validFrom := "2026-04-01"
	invalidTo := "not-a-date"

	_, err := service.GetAll(context.Background(), &validFrom, &invalidTo)

	if err == nil {
		t.Fatal("Expected error for invalid 'to' date, got nil")
	}

	expectedMsg := "invalid to date (expected format: YYYY-MM-DD)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestGetAll_ValidDates_ShouldSucceed(t *testing.T) {
	mockRepo := &mockRepository{
		getAllFunc: func(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error) {
			if from == nil {
				t.Error("Expected 'from' to be parsed")
			}
			if to == nil {
				t.Error("Expected 'to' to be parsed")
			}
			return []Income{}, nil
		},
	}
	service := NewService(mockRepo)

	validFrom := "2026-04-01"
	validTo := "2026-04-30"

	incomes, err := service.GetAll(context.Background(), &validFrom, &validTo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if incomes == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestGetAll_NoDates_ShouldSucceed(t *testing.T) {
	mockRepo := &mockRepository{
		getAllFunc: func(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error) {
			if from != nil {
				t.Error("Expected 'from' to be nil")
			}
			if to != nil {
				t.Error("Expected 'to' to be nil")
			}
			return []Income{
				{ID: 1, OrderID: 123, Amount: 500.00},
			}, nil
		},
	}
	service := NewService(mockRepo)

	incomes, err := service.GetAll(context.Background(), nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(incomes) != 1 {
		t.Errorf("Expected 1 income, got: %d", len(incomes))
	}
}

func TestGetByID_Found(t *testing.T) {
	mockRepo := &mockRepository{
		getByIDFunc: func(ctx context.Context, id int) (*Income, error) {
			return &Income{
				ID:      123,
				OrderID: 456,
				Amount:  750.50,
			}, nil
		},
	}
	service := NewService(mockRepo)

	income, err := service.GetByID(context.Background(), 123)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if income == nil {
		t.Fatal("Expected non-nil income")
	}

	if income.ID != 123 {
		t.Errorf("Expected ID 123, got: %d", income.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getByIDFunc: func(ctx context.Context, id int) (*Income, error) {
			return nil, nil
		},
	}
	service := NewService(mockRepo)

	income, err := service.GetByID(context.Background(), 999)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if income != nil {
		t.Error("Expected nil income for not found")
	}
}

func TestUpdate_Success(t *testing.T) {
	updateCalled := false

	mockRepo := &mockRepository{
		updateFunc: func(ctx context.Context, id int, dto UpdateIncomeDTO) error {
			updateCalled = true
			if id != 123 {
				t.Errorf("Expected ID 123, got: %d", id)
			}
			return nil
		},
	}
	service := NewService(mockRepo)

	newAmount := CustomFloat64(600.00)
	dto := UpdateIncomeDTO{
		Amount: &newAmount,
	}

	err := service.Update(context.Background(), 123, dto)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !updateCalled {
		t.Error("Expected Update to be called on repository")
	}
}

func TestUpdate_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		updateFunc: func(ctx context.Context, id int, dto UpdateIncomeDTO) error {
			return errors.New("database error")
		},
	}
	service := NewService(mockRepo)

	dto := UpdateIncomeDTO{}
	err := service.Update(context.Background(), 123, dto)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("Expected 'database error', got: %s", err.Error())
	}
}

func TestDelete_Success(t *testing.T) {
	deleteCalled := false

	mockRepo := &mockRepository{
		deleteFunc: func(ctx context.Context, id int) error {
			deleteCalled = true
			if id != 123 {
				t.Errorf("Expected ID 123, got: %d", id)
			}
			return nil
		},
	}
	service := NewService(mockRepo)

	err := service.Delete(context.Background(), 123)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !deleteCalled {
		t.Error("Expected Delete to be called on repository")
	}
}

func TestDelete_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		deleteFunc: func(ctx context.Context, id int) error {
			return errors.New("constraint violation")
		},
	}
	service := NewService(mockRepo)

	err := service.Delete(context.Background(), 123)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "constraint violation" {
		t.Errorf("Expected 'constraint violation', got: %s", err.Error())
	}
}
