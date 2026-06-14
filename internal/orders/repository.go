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
	GetAll(ctx context.Context, statuses []OrderStatus, from *time.Time, to *time.Time) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	FinishOrder(ctx context.Context, id int) error
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
	INSERT INTO orders (description, amount_charged, status, estimated_delivery_date, delivery_type, notes, client_name, client_phone)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		*dto.Status,
		dto.EstimatedDeliveryDate,
		*dto.DeliveryType,
		dto.Notes,
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
		client_name, client_phone,
		created_at, updated_at
	FROM orders
	WHERE id = $1;
	`, id)

	var o Order
	err := row.Scan(
		&o.ID, &o.Description, &o.AmountCharged, &o.Status, &o.EntryDate,
		&o.EstimatedDeliveryDate, &o.DeliveryType, &o.Notes,
		&o.ClientName, &o.ClientPhone,
		&o.CreatedAt, &o.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &o, err
}

// -------------------- GET ALL (FILTERS) --------------------

func (r *repository) GetAll(ctx context.Context, statuses []OrderStatus, from *time.Time, to *time.Time) ([]Order, error) {
	query := `
	SELECT 
		id, description, amount_charged, status, entry_date,
		estimated_delivery_date, delivery_type, notes,
		client_name, client_phone,
		created_at, updated_at
	FROM orders
	WHERE 1 = 1
	`

	args := []any{}
	arg := 1

	// Soportar múltiples status con IN clause
	if len(statuses) > 0 {
		if len(statuses) == 1 {
			query += fmt.Sprintf(" AND status = $%d", arg)
			args = append(args, statuses[0])
			arg++
		} else {
			// Construir IN clause: IN ($1, $2, $3)
			placeholders := ""
			for i, status := range statuses {
				if i > 0 {
					placeholders += ", "
				}
				placeholders += fmt.Sprintf("$%d", arg)
				args = append(args, status)
				arg++
			}
			query += fmt.Sprintf(" AND status IN (%s)", placeholders)
		}
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
			&o.ClientName, &o.ClientPhone,
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
		description = $1,
		amount_charged = $2,
		status = $3,
		estimated_delivery_date = $4,
		delivery_type = $5,
		notes = CASE WHEN $6 = '' THEN NULL ELSE $6 END,
		client_name = $7,
		client_phone = CASE WHEN $8 = '' THEN NULL ELSE $8 END,
		updated_at = datetime('now')
	WHERE id = $9;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		dto.Status,
		dto.EstimatedDeliveryDate,
		dto.DeliveryType,
		dto.Notes,
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

// -------------------- FINISH ORDER --------------------
func (r *repository) FinishOrder(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "UPDATE orders SET status = 'completed', updated_at = datetime('now') WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order not found")
	}
	return nil
}
