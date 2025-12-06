package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	Create(ctx context.Context, dto CreateOrderDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Order, error)
	GetAll(ctx context.Context, status *OrderStatus, from *time.Time, to *time.Time) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// -------------------- CREATE --------------------

func (r *repository) Create(ctx context.Context, dto CreateOrderDTO) (int, error) {
	query := `
	INSERT INTO orders (description, amount_charged, status, estimated_delivery_date, delivery_type, notes, paid_50_percent, client_name, client_phone)
	VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, FALSE) , $8, $9)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		dto.Status,
		dto.EstimatedDeliveryDate,
		dto.DeliveryType,
		dto.Notes,
		dto.Paid50Percent,
		dto.ClientName,
		dto.ClientPhone,
	).Scan(&id)

	return id, err
}

// -------------------- GET BY ID --------------------

func (r *repository) GetByID(ctx context.Context, id int) (*Order, error) {
	row := r.db.QueryRowContext(ctx, `
	SELECT 
		id, description, amount_charged, status, entry_date,
		estimated_delivery_date, delivery_type, notes,
		paid_50_percent, client_name, client_phone,
		created_at, updated_at
	FROM orders
	WHERE id = $1;
	`, id)

	var o Order
	err := row.Scan(
		&o.ID, &o.Description, &o.AmountCharged, &o.Status, &o.EntryDate,
		&o.EstimatedDeliveryDate, &o.DeliveryType, &o.Notes,
		&o.Paid50Percent, &o.ClientName, &o.ClientPhone,
		&o.CreatedAt, &o.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &o, err
}

// -------------------- GET ALL (FILTERS) --------------------

func (r *repository) GetAll(ctx context.Context, status *OrderStatus, from *time.Time, to *time.Time) ([]Order, error) {
	query := `
	SELECT 
		id, description, amount_charged, status, entry_date,
		estimated_delivery_date, delivery_type, notes,
		paid_50_percent, client_name, client_phone,
		created_at, updated_at
	FROM orders
	WHERE 1 = 1
	`

	args := []any{}
	arg := 1

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", arg)
		args = append(args, *status)
		arg++
	}

	if from != nil {
		query += fmt.Sprintf(" AND entry_date >= $%d", arg)
		args = append(args, *from)
		arg++
	}

	if to != nil {
		query += fmt.Sprintf(" AND entry_date <= $%d", arg)
		args = append(args, *to)
		arg++
	}

	query += " ORDER BY entry_date DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []Order{}

	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID, &o.Description, &o.AmountCharged, &o.Status, &o.EntryDate,
			&o.EstimatedDeliveryDate, &o.DeliveryType, &o.Notes,
			&o.Paid50Percent, &o.ClientName, &o.ClientPhone,
			&o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

// -------------------- UPDATE --------------------

func (r *repository) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	query := `
	UPDATE orders SET
		description = COALESCE($1, description),
		amount_charged = COALESCE($2, amount_charged),
		status = COALESCE($3, status),
		estimated_delivery_date = COALESCE($4, estimated_delivery_date),
		delivery_type = COALESCE($5, delivery_type),
		notes = COALESCE($6, notes),
		paid_50_percent = COALESCE($7, paid_50_percent),
		client_name = COALESCE($8, client_name),
		client_phone = COALESCE($9, client_phone),
		updated_at = NOW()
	WHERE id = $10;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		dto.Status,
		dto.EstimatedDeliveryDate,
		dto.DeliveryType,
		dto.Notes,
		dto.Paid50Percent,
		dto.ClientName,
		dto.ClientPhone,
		id,
	)

	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order not found")
	}

	return nil
}

// -------------------- DELETE --------------------
// TODO: soft delete?
func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order not found")
	}

	return nil
}
