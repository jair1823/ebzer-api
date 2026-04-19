package expenses

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service interface {
	// Expenses
	CreateExpense(ctx context.Context, dto CreateExpenseDTO) (int, error)
	GetExpenseByID(ctx context.Context, id int) (*Expense, error)
	GetAllExpenses(ctx context.Context, from, to, category, expenseType *string) ([]Expense, error)
	GetExpensesByOrderID(ctx context.Context, orderID int) ([]Expense, error)
	UpdateExpense(ctx context.Context, id int, dto UpdateExpenseDTO) error
	DeleteExpense(ctx context.Context, id int) error

	// Categories
	CreateCategory(ctx context.Context, dto CreateCategoryDTO) (int, error)
	GetCategoryByID(ctx context.Context, id int) (*ExpenseCategory, error)
	GetAllCategories(ctx context.Context) ([]ExpenseCategory, error)
	UpdateCategory(ctx context.Context, id int, dto UpdateCategoryDTO) error
	DeleteCategory(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ---- EXPENSES SERVICE METHODS ----

func (s *service) CreateExpense(ctx context.Context, dto CreateExpenseDTO) (int, error) {
	// Validate amount
	if dto.Amount < 0 {
		return 0, errors.New("amount must be greater than or equal to 0")
	}

	// Validate type
	if dto.Type != string(TypeGeneral) && dto.Type != string(TypeOrderLinked) {
		return 0, errors.New("type must be 'general' or 'order_linked'")
	}

	// Validate order_id consistency with type
	if dto.Type == string(TypeOrderLinked) && dto.OrderID == nil {
		return 0, errors.New("order_id is required when type is 'order_linked'")
	}

	if dto.Type == string(TypeGeneral) && dto.OrderID != nil {
		return 0, errors.New("order_id must be null when type is 'general'")
	}

	// Apply default for date if not provided (matches schema DEFAULT)
	if dto.Date == nil {
		now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
		dto.Date = &now
	}

	return s.repo.CreateExpense(ctx, dto)
}

func (s *service) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	return s.repo.GetExpenseByID(ctx, id)
}

func (s *service) GetAllExpenses(ctx context.Context, from, to, category, expenseType *string) ([]Expense, error) {
	var categoryID *int
	if category != nil {
		// Parse category string to int
		var catID int
		if _, err := fmt.Sscanf(*category, "%d", &catID); err != nil {
			return nil, fmt.Errorf("invalid category: must be a valid integer")
		}
		categoryID = &catID
	}

	return s.repo.GetAllExpenses(ctx, from, to, categoryID, expenseType)
}

func (s *service) GetExpensesByOrderID(ctx context.Context, orderID int) ([]Expense, error) {
	return s.repo.GetExpensesByOrderID(ctx, orderID)
}

func (s *service) UpdateExpense(ctx context.Context, id int, dto UpdateExpenseDTO) error {
	// Validate amount if provided
	if dto.Amount != nil && *dto.Amount < 0 {
		return errors.New("amount must be greater than or equal to 0")
	}

	// Validate type if provided
	if dto.Type != nil {
		if *dto.Type != string(TypeGeneral) && *dto.Type != string(TypeOrderLinked) {
			return errors.New("type must be 'general' or 'order_linked'")
		}
	}

	// If updating type, validate order_id consistency
	if dto.Type != nil {
		if *dto.Type == string(TypeOrderLinked) && dto.OrderID == nil {
			// Check current expense to see if it has order_id
			expense, err := s.repo.GetExpenseByID(ctx, id)
			if err != nil {
				return err
			}
			if expense == nil {
				return errors.New("expense not found")
			}
			if expense.OrderID == nil {
				return errors.New("order_id is required when type is 'order_linked'")
			}
		}

		if *dto.Type == string(TypeGeneral) && dto.OrderID != nil && *dto.OrderID != 0 {
			return errors.New("order_id must be null when type is 'general'")
		}
	}

	return s.repo.UpdateExpense(ctx, id, dto)
}

func (s *service) DeleteExpense(ctx context.Context, id int) error {
	return s.repo.DeleteExpense(ctx, id)
}

// ---- CATEGORIES SERVICE METHODS ----

func (s *service) CreateCategory(ctx context.Context, dto CreateCategoryDTO) (int, error) {
	// Validate name
	if dto.Name == "" {
		return 0, errors.New("name is required")
	}

	return s.repo.CreateCategory(ctx, dto)
}

func (s *service) GetCategoryByID(ctx context.Context, id int) (*ExpenseCategory, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *service) GetAllCategories(ctx context.Context) ([]ExpenseCategory, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *service) UpdateCategory(ctx context.Context, id int, dto UpdateCategoryDTO) error {
	// Validate name if provided
	if dto.Name != nil && *dto.Name == "" {
		return errors.New("name cannot be empty")
	}

	return s.repo.UpdateCategory(ctx, id, dto)
}

func (s *service) DeleteCategory(ctx context.Context, id int) error {
	return s.repo.DeleteCategory(ctx, id)
}
