package incomes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	Create(ctx context.Context, dto CreateIncomeDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Income, error)
	GetAll(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error)
	GetByOrderID(ctx context.Context, orderID int) ([]Income, error)
	Update(ctx context.Context, id int, dto UpdateIncomeDTO) error
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// -------------------- CREATE --------------------

func (r *repository) Create(ctx context.Context, dto CreateIncomeDTO) (int, error) {
	query := `
	INSERT INTO income (order_id, amount, date)
	VALUES ($1, $2, datetime('now'))
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.OrderID,
		dto.Amount,
	).Scan(&id)

	return id, err
}

// -------------------- GET BY ID --------------------

func (r *repository) GetByID(ctx context.Context, id int) (*Income, error) {
	row := r.db.QueryRowContext(ctx, `
	SELECT 
		id, order_id, amount, date,
		created_at, updated_at
	FROM income
	WHERE id = $1;
	`, id)

	var o Income
	err := row.Scan(
		&o.ID, &o.OrderID, &o.Amount, &o.Date,
		&o.CreatedAt, &o.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &o, err
}

// -------------------- GET ALL (FILTERS) --------------------

func (r *repository) GetAll(ctx context.Context, from *time.Time, to *time.Time) ([]Income, error) {
	query := `
	SELECT 
		id, order_id, amount, date,
		created_at, updated_at
	FROM income
	WHERE 1 = 1
	`

	args := []any{}
	arg := 1

	if from != nil {
		query += fmt.Sprintf(" AND date >= $%d", arg)
		args = append(args, *from)
		arg++
	}

	if to != nil {
		query += fmt.Sprintf(" AND date <= $%d", arg)
		args = append(args, *to)
		arg++
	}

	query += " ORDER BY date DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	incomes := []Income{}

	for rows.Next() {
		var i Income
		err := rows.Scan(
			&i.ID, &i.OrderID, &i.Amount, &i.Date,
			&i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		incomes = append(incomes, i)
	}

	return incomes, nil
}

// -------------------- GET BY ORDER ID --------------------

func (r *repository) GetByOrderID(ctx context.Context, orderID int) ([]Income, error) {
	query := `
	SELECT 
		id, order_id, amount, date,
		created_at, updated_at
	FROM income
	WHERE order_id = $1
	ORDER BY date DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	incomes := []Income{}

	for rows.Next() {
		var i Income
		err := rows.Scan(
			&i.ID, &i.OrderID, &i.Amount, &i.Date,
			&i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		incomes = append(incomes, i)
	}

	return incomes, nil
}

// -------------------- UPDATE --------------------

func (r *repository) Update(ctx context.Context, id int, dto UpdateIncomeDTO) error {
	query := `
	UPDATE income SET
		order_id = COALESCE($1, order_id),
		amount = COALESCE($2, amount),
		updated_at = datetime('now')
	WHERE id = $3;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.OrderID,
		dto.Amount,
		id,
	)

	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("income not found")
	}

	return nil
}

// -------------------- DELETE --------------------
// TODO: soft delete?
func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM income WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("income not found")
	}

	return nil
}
