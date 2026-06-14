package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Repository interface {
	Create(ctx context.Context, dto CreateExpenseDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Expense, error)
	GetAll(ctx context.Context, from, to *string, comercioID *int) ([]Expense, error)
	Update(ctx context.Context, id int, dto UpdateExpenseDTO) error
	Delete(ctx context.Context, id int) error
	ComercioExists(ctx context.Context, id int) (bool, error)
	CreateComercio(ctx context.Context, dto CreateComercioDTO) (int, error)
	GetComercios(ctx context.Context) ([]Comercio, error)
	UpdateComercio(ctx context.Context, id int, dto UpdateComercioDTO) error
	DeleteComercio(ctx context.Context, id int) error
	CreateProduct(ctx context.Context, dto CreateProductDTO) (int, error)
	GetProducts(ctx context.Context, comercioID *int) ([]Product, error)
	UpdateProduct(ctx context.Context, id int, dto UpdateProductDTO) error
	DeleteProduct(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, dto CreateExpenseDTO) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO expenses (comercio_id, date, description)
		VALUES ($1, COALESCE($2, date('now')), $3)
		RETURNING id;
	`, dto.ComercioID, dto.Date, dto.Description).Scan(&id); err != nil {
		return 0, err
	}

	if err := r.replaceItems(ctx, tx, id, dto.ComercioID, dto.Items); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) GetByID(ctx context.Context, id int) (*Expense, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			e.id, e.comercio_id, e.description, e.date, e.created_at,
			COALESCE(SUM(ei.quantity * ei.unit_price), 0) AS total,
			c.id, c.name, c.description, c.created_at
		FROM expenses e
		INNER JOIN comercios c ON c.id = e.comercio_id
		LEFT JOIN expense_items ei ON ei.expense_id = e.id
		WHERE e.id = $1
		GROUP BY e.id, c.id;
	`, id)

	expense, err := scanExpense(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getItemsForExpenseIDs(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	expense.Items = items[id]

	return expense, nil
}

func (r *repository) GetAll(ctx context.Context, from, to *string, comercioID *int) ([]Expense, error) {
	query := `
		SELECT
			e.id, e.comercio_id, e.description, e.date, e.created_at,
			COALESCE(SUM(ei.quantity * ei.unit_price), 0) AS total,
			c.id, c.name, c.description, c.created_at
		FROM expenses e
		INNER JOIN comercios c ON c.id = e.comercio_id
		LEFT JOIN expense_items ei ON ei.expense_id = e.id
		WHERE 1 = 1
	`

	args := []any{}
	arg := 1

	if from != nil {
		query += fmt.Sprintf(" AND date(e.date) >= date($%d)", arg)
		args = append(args, *from)
		arg++
	}

	if to != nil {
		query += fmt.Sprintf(" AND date(e.date) <= date($%d)", arg)
		args = append(args, *to)
		arg++
	}

	if comercioID != nil {
		query += fmt.Sprintf(" AND e.comercio_id = $%d", arg)
		args = append(args, *comercioID)
	}

	query += " GROUP BY e.id, c.id ORDER BY date(e.date) DESC, e.id DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []Expense{}
	ids := []int{}
	for rows.Next() {
		expense, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, *expense)
		ids = append(ids, expense.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	itemsByExpenseID, err := r.getItemsForExpenseIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for index := range expenses {
		expenses[index].Items = itemsByExpenseID[expenses[index].ID]
	}

	return expenses, nil
}

func (r *repository) Update(ctx context.Context, id int, dto UpdateExpenseDTO) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE expenses
		SET comercio_id = $1,
			date = COALESCE($2, date),
			description = $3
		WHERE id = $4;
	`, dto.ComercioID, dto.Date, dto.Description, id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("expense not found")
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM expense_items WHERE expense_id = $1", id); err != nil {
		return err
	}

	if err := r.replaceItems(ctx, tx, id, dto.ComercioID, dto.Items); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM expenses WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("expense not found")
	}

	return nil
}

func (r *repository) ComercioExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM comercios WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (r *repository) CreateComercio(ctx context.Context, dto CreateComercioDTO) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO comercios (name, description)
		VALUES ($1, $2)
		RETURNING id;
	`, dto.Name, dto.Description).Scan(&id)
	return id, err
}

func (r *repository) GetComercios(ctx context.Context) ([]Comercio, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, created_at
		FROM comercios
		ORDER BY name ASC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comercios := []Comercio{}
	for rows.Next() {
		comercio, err := scanComercio(rows)
		if err != nil {
			return nil, err
		}
		comercios = append(comercios, *comercio)
	}

	return comercios, rows.Err()
}

func (r *repository) UpdateComercio(ctx context.Context, id int, dto UpdateComercioDTO) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE comercios
		SET name = COALESCE($1, name),
			description = $2
		WHERE id = $3;
	`, dto.Name, dto.Description, id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("comercio not found")
	}
	return nil
}

func (r *repository) DeleteComercio(ctx context.Context, id int) error {
	var usageCount int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM expenses WHERE comercio_id = $1", id).Scan(&usageCount); err != nil {
		return err
	}
	if usageCount > 0 {
		return errors.New("comercio is in use")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM comercios WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("comercio not found")
	}
	return nil
}

func (r *repository) CreateProduct(ctx context.Context, dto CreateProductDTO) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO products (comercio_id, name, default_price)
		VALUES ($1, $2, $3)
		RETURNING id;
	`, dto.ComercioID, dto.Name, dto.DefaultPrice).Scan(&id)
	return id, err
}

func (r *repository) GetProducts(ctx context.Context, comercioID *int) ([]Product, error) {
	query := `
		SELECT
			p.id, p.comercio_id, p.name, p.default_price, p.created_at,
			c.id, c.name, c.description, c.created_at
		FROM products p
		INNER JOIN comercios c ON c.id = p.comercio_id
		WHERE 1 = 1
	`
	args := []any{}
	if comercioID != nil {
		query += " AND p.comercio_id = $1"
		args = append(args, *comercioID)
	}
	query += " ORDER BY p.name ASC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *product)
	}

	return products, rows.Err()
}

func (r *repository) UpdateProduct(ctx context.Context, id int, dto UpdateProductDTO) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE products
		SET comercio_id = COALESCE($1, comercio_id),
			name = COALESCE($2, name),
			default_price = COALESCE($3, default_price)
		WHERE id = $4;
	`, dto.ComercioID, dto.Name, dto.DefaultPrice, id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (r *repository) DeleteProduct(ctx context.Context, id int) error {
	var usageCount int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM expense_items WHERE product_id = $1", id).Scan(&usageCount); err != nil {
		return err
	}
	if usageCount > 0 {
		return errors.New("product is in use")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (r *repository) replaceItems(ctx context.Context, tx *sql.Tx, expenseID int, comercioID int, items []CreateExpenseItemDTO) error {
	for _, item := range items {
		productID, productName, err := resolveExpenseProduct(ctx, tx, comercioID, item)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO expense_items (expense_id, product_id, product_name, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5);
		`, expenseID, productID, productName, item.Quantity, item.UnitPrice); err != nil {
			return err
		}
	}
	return nil
}

func resolveExpenseProduct(ctx context.Context, tx *sql.Tx, comercioID int, item CreateExpenseItemDTO) (int, string, error) {
	productName := strings.TrimSpace(item.ProductName)
	unitPrice := float64(item.UnitPrice)

	if item.ProductID != nil && *item.ProductID > 0 {
		var id int
		var name string
		var defaultPrice float64
		err := tx.QueryRowContext(ctx, `
			SELECT id, name, default_price
			FROM products
			WHERE id = $1 AND comercio_id = $2;
		`, *item.ProductID, comercioID).Scan(&id, &name, &defaultPrice)
		if err == sql.ErrNoRows {
			return 0, "", errors.New("product not found for comercio")
		}
		if err != nil {
			return 0, "", err
		}
		if defaultPrice != unitPrice {
			if _, err := tx.ExecContext(ctx, "UPDATE products SET default_price = $1 WHERE id = $2", unitPrice, id); err != nil {
				return 0, "", err
			}
		}
		return id, name, nil
	}

	var id int
	var name string
	var defaultPrice float64
	err := tx.QueryRowContext(ctx, `
		SELECT id, name, default_price
		FROM products
		WHERE comercio_id = $1 AND name = $2 COLLATE NOCASE;
	`, comercioID, productName).Scan(&id, &name, &defaultPrice)
	if err == nil {
		if defaultPrice != unitPrice {
			if _, err := tx.ExecContext(ctx, "UPDATE products SET default_price = $1 WHERE id = $2", unitPrice, id); err != nil {
				return 0, "", err
			}
		}
		return id, name, nil
	}
	if err != sql.ErrNoRows {
		return 0, "", err
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO products (comercio_id, name, default_price)
		VALUES ($1, $2, $3)
		RETURNING id, name;
	`, comercioID, productName, unitPrice).Scan(&id, &name)
	if err != nil {
		return 0, "", err
	}

	return id, name, nil
}

func (r *repository) getItemsForExpenseIDs(ctx context.Context, expenseIDs []int) (map[int][]ExpenseItem, error) {
	itemsByExpenseID := make(map[int][]ExpenseItem, len(expenseIDs))
	if len(expenseIDs) == 0 {
		return itemsByExpenseID, nil
	}

	placeholders := make([]string, len(expenseIDs))
	args := make([]any, len(expenseIDs))
	for index, id := range expenseIDs {
		placeholders[index] = fmt.Sprintf("$%d", index+1)
		args[index] = id
		itemsByExpenseID[id] = []ExpenseItem{}
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, expense_id, product_id, product_name, quantity, unit_price, quantity * unit_price AS line_total
		FROM expense_items
		WHERE expense_id IN (%s)
		ORDER BY id ASC;
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ExpenseItem
		if err := rows.Scan(
			&item.ID,
			&item.ExpenseID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.LineTotal,
		); err != nil {
			return nil, err
		}
		itemsByExpenseID[item.ExpenseID] = append(itemsByExpenseID[item.ExpenseID], item)
	}

	return itemsByExpenseID, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanExpense(row rowScanner) (*Expense, error) {
	var expense Expense
	var description sql.NullString
	var comercio Comercio
	var comercioDescription sql.NullString

	if err := row.Scan(
		&expense.ID,
		&expense.ComercioID,
		&description,
		&expense.Date,
		&expense.CreatedAt,
		&expense.Total,
		&comercio.ID,
		&comercio.Name,
		&comercioDescription,
		&comercio.CreatedAt,
	); err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		expense.Description = &value
	}
	if comercioDescription.Valid {
		value := comercioDescription.String
		comercio.Description = &value
	}
	expense.Comercio = &comercio

	return &expense, nil
}

func scanComercio(row rowScanner) (*Comercio, error) {
	var comercio Comercio
	var description sql.NullString

	if err := row.Scan(
		&comercio.ID,
		&comercio.Name,
		&description,
		&comercio.CreatedAt,
	); err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		comercio.Description = &value
	}

	return &comercio, nil
}

func scanProduct(row rowScanner) (*Product, error) {
	var product Product
	var comercio Comercio
	var comercioDescription sql.NullString

	if err := row.Scan(
		&product.ID,
		&product.ComercioID,
		&product.Name,
		&product.DefaultPrice,
		&product.CreatedAt,
		&comercio.ID,
		&comercio.Name,
		&comercioDescription,
		&comercio.CreatedAt,
	); err != nil {
		return nil, err
	}

	if comercioDescription.Valid {
		value := comercioDescription.String
		comercio.Description = &value
	}
	product.Comercio = &comercio

	return &product, nil
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "unique")
}
