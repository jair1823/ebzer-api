package expenses

import (
	"context"
	"errors"
	"testing"
)

// Mock repository para testing
type mockRepository struct {
	createExpenseFunc  func(ctx context.Context, dto CreateExpenseDTO) (int, error)
	getExpenseByIDFunc func(ctx context.Context, id int) (*Expense, error)
	getAllExpensesFunc func(ctx context.Context, from, to *string, categoryID *int, expenseType *string) ([]Expense, error)
	updateExpenseFunc  func(ctx context.Context, id int, dto UpdateExpenseDTO) error
	deleteExpenseFunc  func(ctx context.Context, id int) error

	createCategoryFunc   func(ctx context.Context, dto CreateCategoryDTO) (int, error)
	getCategoryByIDFunc  func(ctx context.Context, id int) (*ExpenseCategory, error)
	getAllCategoriesFunc func(ctx context.Context) ([]ExpenseCategory, error)
	updateCategoryFunc   func(ctx context.Context, id int, dto UpdateCategoryDTO) error
	deleteCategoryFunc   func(ctx context.Context, id int) error
}

// Expenses mock implementations
func (m *mockRepository) CreateExpense(ctx context.Context, dto CreateExpenseDTO) (int, error) {
	if m.createExpenseFunc != nil {
		return m.createExpenseFunc(ctx, dto)
	}
	return 1, nil
}

func (m *mockRepository) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	if m.getExpenseByIDFunc != nil {
		return m.getExpenseByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetAllExpenses(ctx context.Context, from, to *string, categoryID *int, expenseType *string) ([]Expense, error) {
	if m.getAllExpensesFunc != nil {
		return m.getAllExpensesFunc(ctx, from, to, categoryID, expenseType)
	}
	return []Expense{}, nil
}

func (m *mockRepository) GetExpensesByOrderID(ctx context.Context, orderID int) ([]Expense, error) {
	return []Expense{}, nil
}

func (m *mockRepository) UpdateExpense(ctx context.Context, id int, dto UpdateExpenseDTO) error {
	if m.updateExpenseFunc != nil {
		return m.updateExpenseFunc(ctx, id, dto)
	}
	return nil
}

func (m *mockRepository) DeleteExpense(ctx context.Context, id int) error {
	if m.deleteExpenseFunc != nil {
		return m.deleteExpenseFunc(ctx, id)
	}
	return nil
}

// Categories mock implementations
func (m *mockRepository) CreateCategory(ctx context.Context, dto CreateCategoryDTO) (int, error) {
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(ctx, dto)
	}
	return 1, nil
}

func (m *mockRepository) GetCategoryByID(ctx context.Context, id int) (*ExpenseCategory, error) {
	if m.getCategoryByIDFunc != nil {
		return m.getCategoryByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetAllCategories(ctx context.Context) ([]ExpenseCategory, error) {
	if m.getAllCategoriesFunc != nil {
		return m.getAllCategoriesFunc(ctx)
	}
	return []ExpenseCategory{}, nil
}

func (m *mockRepository) UpdateCategory(ctx context.Context, id int, dto UpdateCategoryDTO) error {
	if m.updateCategoryFunc != nil {
		return m.updateCategoryFunc(ctx, id, dto)
	}
	return nil
}

func (m *mockRepository) DeleteCategory(ctx context.Context, id int) error {
	if m.deleteCategoryFunc != nil {
		return m.deleteCategoryFunc(ctx, id)
	}
	return nil
}

// ================== EXPENSES TESTS ==================

func TestCreateExpense_AppliesDefaultDate(t *testing.T) {
	var capturedDTO CreateExpenseDTO

	mockRepo := &mockRepository{
		createExpenseFunc: func(ctx context.Context, dto CreateExpenseDTO) (int, error) {
			capturedDTO = dto
			return 1, nil
		},
	}

	service := NewService(mockRepo)

	// Crear expense sin fecha
	dto := CreateExpenseDTO{
		Description: "Test expense",
		Amount:      100.00,
		Type:        string(TypeGeneral),
		Date:        nil, // No se proporciona fecha
	}

	_, err := service.CreateExpense(context.Background(), dto)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verificar que se aplicó un default
	if capturedDTO.Date == nil {
		t.Error("Expected Date to be set with default value, got nil")
	}
}

func TestCreateExpense_ValidatesAmount(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	tests := []struct {
		name        string
		amount      CustomFloat64
		expectError bool
	}{
		{"valid amount zero", 0, false},
		{"valid amount positive", 100.50, false},
		{"invalid amount negative", -50.00, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := CreateExpenseDTO{
				Description: "Test",
				Amount:      tt.amount,
				Type:        string(TypeGeneral),
			}

			_, err := service.CreateExpense(context.Background(), dto)

			if tt.expectError && err == nil {
				t.Error("Expected error for negative amount, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestCreateExpense_ValidatesTypeAndOrderID(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	tests := []struct {
		name        string
		expenseType string
		orderID     *int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "general with nil order_id - valid",
			expenseType: string(TypeGeneral),
			orderID:     nil,
			expectError: false,
		},
		{
			name:        "general with order_id - invalid",
			expenseType: string(TypeGeneral),
			orderID:     intPtr(123),
			expectError: true,
			errorMsg:    "order_id must be null when type is 'general'",
		},
		{
			name:        "order_linked with order_id - valid",
			expenseType: string(TypeOrderLinked),
			orderID:     intPtr(123),
			expectError: false,
		},
		{
			name:        "order_linked with nil order_id - invalid",
			expenseType: string(TypeOrderLinked),
			orderID:     nil,
			expectError: true,
			errorMsg:    "order_id is required when type is 'order_linked'",
		},
		{
			name:        "invalid type",
			expenseType: "invalid",
			orderID:     nil,
			expectError: true,
			errorMsg:    "type must be 'general' or 'order_linked'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := CreateExpenseDTO{
				Description: "Test",
				Amount:      100.00,
				Type:        tt.expenseType,
				OrderID:     tt.orderID,
			}

			_, err := service.CreateExpense(context.Background(), dto)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error '%s', got nil", tt.errorMsg)
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestUpdateExpense_ValidatesAmount(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	negativeAmount := CustomFloat64(-10.00)
	validAmount := CustomFloat64(50.00)

	tests := []struct {
		name        string
		amount      *CustomFloat64
		expectError bool
	}{
		{"nil amount - valid", nil, false},
		{"positive amount - valid", &validAmount, false},
		{"negative amount - invalid", &negativeAmount, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := UpdateExpenseDTO{
				Amount: tt.amount,
			}

			err := service.UpdateExpense(context.Background(), 1, dto)

			if tt.expectError && err == nil {
				t.Error("Expected error for negative amount, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestUpdateExpense_ValidatesTypeChange(t *testing.T) {
	mockRepo := &mockRepository{
		getExpenseByIDFunc: func(ctx context.Context, id int) (*Expense, error) {
			return &Expense{
				ID:      1,
				OrderID: intPtr(123),
				Type:    TypeOrderLinked,
			}, nil
		},
	}
	service := NewService(mockRepo)

	tests := []struct {
		name        string
		newType     *string
		orderID     *int
		expectError bool
	}{
		{
			name:        "nil type - valid",
			newType:     nil,
			orderID:     nil,
			expectError: false,
		},
		{
			name:        "valid type no change",
			newType:     stringPtr(string(TypeOrderLinked)),
			orderID:     intPtr(123),
			expectError: false,
		},
		{
			name:        "invalid type",
			newType:     stringPtr("invalid"),
			orderID:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := UpdateExpenseDTO{
				Type:    tt.newType,
				OrderID: tt.orderID,
			}

			err := service.UpdateExpense(context.Background(), 1, dto)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestGetAllExpenses_ValidatesCategoryFilter(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	tests := []struct {
		name        string
		category    *string
		expectError bool
	}{
		{
			name:        "nil category - valid",
			category:    nil,
			expectError: false,
		},
		{
			name:        "valid category id",
			category:    stringPtr("123"),
			expectError: false,
		},
		{
			name:        "invalid category - not a number",
			category:    stringPtr("invalid"),
			expectError: true,
		},
		{
			name:        "invalid category - empty string",
			category:    stringPtr(""),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetAllExpenses(context.Background(), nil, nil, tt.category, nil)

			if tt.expectError && err == nil {
				t.Error("Expected error for invalid category, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// ================== CATEGORIES TESTS ==================

func TestCreateCategory_ValidatesName(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	tests := []struct {
		name         string
		categoryName string
		expectError  bool
	}{
		{"valid name", "Materials", false},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := CreateCategoryDTO{
				Name: tt.categoryName,
			}

			_, err := service.CreateCategory(context.Background(), dto)

			if tt.expectError && err == nil {
				t.Error("Expected error for empty name, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestUpdateCategory_ValidatesName(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	validName := "Updated Category"
	emptyName := ""

	tests := []struct {
		name         string
		categoryName *string
		expectError  bool
	}{
		{"nil name - valid", nil, false},
		{"valid name", &validName, false},
		{"empty name - invalid", &emptyName, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := UpdateCategoryDTO{
				Name: tt.categoryName,
			}

			err := service.UpdateCategory(context.Background(), 1, dto)

			if tt.expectError && err == nil {
				t.Error("Expected error for empty name, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestDeleteExpense_CallsRepository(t *testing.T) {
	called := false
	mockRepo := &mockRepository{
		deleteExpenseFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 123 {
				t.Errorf("Expected id 123, got %d", id)
			}
			return nil
		},
	}

	service := NewService(mockRepo)
	err := service.DeleteExpense(context.Background(), 123)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Expected repository DeleteExpense to be called")
	}
}

func TestDeleteCategory_CallsRepository(t *testing.T) {
	called := false
	mockRepo := &mockRepository{
		deleteCategoryFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 456 {
				t.Errorf("Expected id 456, got %d", id)
			}
			return nil
		},
	}

	service := NewService(mockRepo)
	err := service.DeleteCategory(context.Background(), 456)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Expected repository DeleteCategory to be called")
	}
}

func TestGetExpenseByID_ReturnsExpense(t *testing.T) {
	expectedExpense := &Expense{
		ID:          1,
		Description: "Test Expense",
		Amount:      100.00,
	}

	mockRepo := &mockRepository{
		getExpenseByIDFunc: func(ctx context.Context, id int) (*Expense, error) {
			if id == 1 {
				return expectedExpense, nil
			}
			return nil, errors.New("not found")
		},
	}

	service := NewService(mockRepo)
	result, err := service.GetExpenseByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != expectedExpense {
		t.Error("Expected returned expense to match")
	}
}

func TestGetCategoryByID_ReturnsCategory(t *testing.T) {
	expectedCategory := &ExpenseCategory{
		ID:   1,
		Name: "Materials",
	}

	mockRepo := &mockRepository{
		getCategoryByIDFunc: func(ctx context.Context, id int) (*ExpenseCategory, error) {
			if id == 1 {
				return expectedCategory, nil
			}
			return nil, errors.New("not found")
		},
	}

	service := NewService(mockRepo)
	result, err := service.GetCategoryByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != expectedCategory {
		t.Error("Expected returned category to match")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
