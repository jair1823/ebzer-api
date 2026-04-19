package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	// Expenses operations
	CreateExpense(ctx context.Context, dto CreateExpenseDTO) (int, error)
	GetExpenseByID(ctx context.Context, id int) (*Expense, error)
	GetAllExpenses(ctx context.Context, from, to *string, categoryID *int, expenseType *string) ([]Expense, error)
	GetExpensesByOrderID(ctx context.Context, orderID int) ([]Expense, error)
	UpdateExpense(ctx context.Context, id int, dto UpdateExpenseDTO) error
	DeleteExpense(ctx context.Context, id int) error

	// Categories operations
	CreateCategory(ctx context.Context, dto CreateCategoryDTO) (int, error)
	GetCategoryByID(ctx context.Context, id int) (*ExpenseCategory, error)
	GetAllCategories(ctx context.Context) ([]ExpenseCategory, error)
	UpdateCategory(ctx context.Context, id int, dto UpdateCategoryDTO) error
	DeleteCategory(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// ---- EXPENSES REPOSITORY METHODS ----

func (r *repository) CreateExpense(ctx context.Context, dto CreateExpenseDTO) (int, error) {
	query := `
	INSERT INTO expenses (description, amount, date, order_id, category_id, type)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Description,
		dto.Amount,
		dto.Date,
		dto.OrderID,
		dto.CategoryID,
		dto.Type,
	).Scan(&id)

	return id, err
}

func (r *repository) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	row := r.db.QueryRowContext(ctx, `
	SELECT 
		id, description, amount, date, order_id, category_id, type, created_at
	FROM expenses
	WHERE id = $1;
	`, id)

	var e Expense
	err := row.Scan(
		&e.ID, &e.Description, &e.Amount, &e.Date,
		&e.OrderID, &e.CategoryID, &e.Type, &e.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &e, err
}

func (r *repository) GetAllExpenses(ctx context.Context, from, to *string, categoryID *int, expenseType *string) ([]Expense, error) {
	query := `
	SELECT 
		id, description, amount, date, order_id, category_id, type, created_at
	FROM expenses
	WHERE 1 = 1
	`

	args := []any{}
	argCount := 1

	if from != nil {
		query += fmt.Sprintf(" AND date >= $%d", argCount)
		args = append(args, *from)
		argCount++
	}

	if to != nil {
		query += fmt.Sprintf(" AND date <= $%d", argCount)
		args = append(args, *to)
		argCount++
	}

	if categoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if expenseType != nil {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *expenseType)
		argCount++
	}

	query += " ORDER BY date DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []Expense{}
	for rows.Next() {
		var e Expense
		err := rows.Scan(
			&e.ID, &e.Description, &e.Amount, &e.Date,
			&e.OrderID, &e.CategoryID, &e.Type, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (r *repository) GetExpensesByOrderID(ctx context.Context, orderID int) ([]Expense, error) {
	query := `
	SELECT 
		id, description, amount, date, order_id, category_id, type, created_at
	FROM expenses
	WHERE order_id = $1
	ORDER BY date DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []Expense{}
	for rows.Next() {
		var e Expense
		err := rows.Scan(
			&e.ID, &e.Description, &e.Amount, &e.Date,
			&e.OrderID, &e.CategoryID, &e.Type, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (r *repository) UpdateExpense(ctx context.Context, id int, dto UpdateExpenseDTO) error {
	query := `
	UPDATE expenses SET
		description = $1,
		amount = $2,
		date = $3,
		order_id = $4,
		category_id = $5,
		type = $6
	WHERE id = $7;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.Description,
		dto.Amount,
		dto.Date,
		dto.OrderID,
		dto.CategoryID,
		dto.Type,
		id,
	)

	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("expense not found")
	}

	return nil
}

func (r *repository) DeleteExpense(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM expenses WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("expense not found")
	}

	return nil
}

// ---- CATEGORIES REPOSITORY METHODS ----

func (r *repository) CreateCategory(ctx context.Context, dto CreateCategoryDTO) (int, error) {
	query := `
	INSERT INTO expense_categories (name, description)
	VALUES ($1, $2)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query, dto.Name, dto.Description).Scan(&id)

	return id, err
}

func (r *repository) GetCategoryByID(ctx context.Context, id int) (*ExpenseCategory, error) {
	row := r.db.QueryRowContext(ctx, `
	SELECT id, name, description, created_at
	FROM expense_categories
	WHERE id = $1;
	`, id)

	var c ExpenseCategory
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &c, err
}

func (r *repository) GetAllCategories(ctx context.Context) ([]ExpenseCategory, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, name, description, created_at
	FROM expense_categories
	ORDER BY name ASC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []ExpenseCategory{}
	for rows.Next() {
		var c ExpenseCategory
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (r *repository) UpdateCategory(ctx context.Context, id int, dto UpdateCategoryDTO) error {
	query := `
	UPDATE expense_categories SET
		name = COALESCE($1, name),
		description = COALESCE($2, description)
	WHERE id = $3;
	`

	result, err := r.db.ExecContext(ctx, query, dto.Name, dto.Description, id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("category not found")
	}

	return nil
}

func (r *repository) DeleteCategory(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM expense_categories WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("category not found")
	}

	return nil
}
