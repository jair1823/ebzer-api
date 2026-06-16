package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	internaldb "creaciones-api/internal/db"
)

type Repository interface {
	Create(ctx context.Context, dto CreateOrderDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Order, error)
	GetAll(ctx context.Context, filter OrderFilter) ([]Order, error)
	Update(ctx context.Context, id int, dto UpdateOrderDTO) error
	FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error)
	Delete(ctx context.Context, id int, actorID *int) error
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
	INSERT INTO orders (description, amount_charged, status_id, estimated_delivery_date, delivery_type, platform, notes, client_name, client_phone)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		dto.StatusID,
		dto.EstimatedDeliveryDate,
		dto.DeliveryType,
		dto.Platform,
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
		o.id, o.description, o.amount_charged, o.status_id, o.entry_date,
		o.estimated_delivery_date, o.delivery_type, o.platform, o.notes,
		o.client_name, o.client_phone,
		o.paid_at, o.created_at, o.updated_at,
		s.id, s.name, s.display_name, s.color, s.order_position,
		s.is_system_status, s.is_final_status, s.is_active, s.created_at, s.updated_at,
		COALESCE(p.total_paid, 0)
	FROM orders o
	LEFT JOIN order_statuses s ON s.id = o.status_id
	LEFT JOIN (
		SELECT order_id, SUM(amount) AS total_paid
		FROM income
		WHERE deleted_at IS NULL
		GROUP BY order_id
	) p ON p.order_id = o.id
	WHERE o.id = $1 AND o.deleted_at IS NULL;
	`, id)

	var o Order
	var status nullableOrderStatus
	var totalPaid float64
	err := row.Scan(
		&o.ID, &o.Description, &o.AmountCharged, &o.StatusID, &o.EntryDate,
		&o.EstimatedDeliveryDate, &o.DeliveryType, &o.Platform, &o.Notes,
		&o.ClientName, &o.ClientPhone,
		&o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
		&status.ID, &status.Name, &status.DisplayName, &status.Color,
		&status.OrderPosition, &status.IsSystemStatus, &status.IsFinalStatus,
		&status.IsActive, &status.CreatedAt, &status.UpdatedAt,
		&totalPaid,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	o.Status = status.toOrderStatus()
	o.PaymentStatus = paymentStatusFromTotals(o.AmountCharged, totalPaid)

	return &o, nil
}

// -------------------- GET ALL (FILTERS) --------------------

func (r *repository) GetAll(ctx context.Context, filter OrderFilter) ([]Order, error) {
	query := `
	SELECT 
		o.id, o.description, o.amount_charged, o.status_id, o.entry_date,
		o.estimated_delivery_date, o.delivery_type, o.platform, o.notes,
		o.client_name, o.client_phone,
		o.paid_at, o.created_at, o.updated_at,
		s.id, s.name, s.display_name, s.color, s.order_position,
		s.is_system_status, s.is_final_status, s.is_active, s.created_at, s.updated_at,
		COALESCE(p.total_paid, 0)
	FROM orders o
	LEFT JOIN order_statuses s ON s.id = o.status_id
	LEFT JOIN (
		SELECT order_id, SUM(amount) AS total_paid
		FROM income
		WHERE deleted_at IS NULL
		GROUP BY order_id
	) p ON p.order_id = o.id
	WHERE o.deleted_at IS NULL
	`

	args := []any{}
	arg := 1

	if filter.StatusID != nil {
		query += fmt.Sprintf(" AND o.status_id = $%d", arg)
		args = append(args, *filter.StatusID)
		arg++
	}

	if len(filter.StatusIDs) > 0 {
		placeholders := make([]string, 0, len(filter.StatusIDs))
		for _, statusID := range filter.StatusIDs {
			placeholders = append(placeholders, fmt.Sprintf("$%d", arg))
			args = append(args, statusID)
			arg++
		}
		query += fmt.Sprintf(" AND o.status_id IN (%s)", strings.Join(placeholders, ", "))
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND o.entry_date >= $%d", arg)
		args = append(args, *filter.From)
		arg++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND o.entry_date <= $%d", arg)
		args = append(args, *filter.To)
		arg++
	}

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		term := "%" + strings.ToLower(strings.TrimSpace(*filter.Search)) + "%"
		query += fmt.Sprintf(
			" AND (LOWER(o.description) LIKE $%d OR LOWER(COALESCE(o.notes, '')) LIKE $%d OR LOWER(COALESCE(o.client_name, '')) LIKE $%d OR LOWER(COALESCE(o.client_phone, '')) LIKE $%d)",
			arg, arg+1, arg+2, arg+3,
		)
		args = append(args, term, term, term, term)
		arg += 4
	}

	if filter.DeliveryFrom != nil {
		query += fmt.Sprintf(" AND date(o.estimated_delivery_date) >= date($%d)", arg)
		args = append(args, *filter.DeliveryFrom)
		arg++
	}

	if filter.DeliveryTo != nil {
		query += fmt.Sprintf(" AND date(o.estimated_delivery_date) <= date($%d)", arg)
		args = append(args, *filter.DeliveryTo)
		arg++
	}

	if filter.Platform != nil {
		query += fmt.Sprintf(" AND o.platform = $%d", arg)
		args = append(args, *filter.Platform)
		arg++
	}

	if filter.AmountMin != nil {
		query += fmt.Sprintf(" AND o.amount_charged >= $%d", arg)
		args = append(args, *filter.AmountMin)
		arg++
	}

	if filter.AmountMax != nil {
		query += fmt.Sprintf(" AND o.amount_charged <= $%d", arg)
		args = append(args, *filter.AmountMax)
		arg++
	}

	if filter.PaymentStatus != nil {
		switch *filter.PaymentStatus {
		case PaymentStatusFilterUnpaid:
			query += " AND COALESCE(p.total_paid, 0) <= 0"
		case PaymentStatusFilterPartial:
			query += " AND COALESCE(p.total_paid, 0) > 0 AND COALESCE(p.total_paid, 0) < o.amount_charged"
		case PaymentStatusFilterPaid:
			query += " AND COALESCE(p.total_paid, 0) >= o.amount_charged AND COALESCE(p.total_paid, 0) > 0"
		}
	}

	if filter.Overdue != nil && *filter.Overdue {
		today := filter.Today
		if today.IsZero() {
			today = time.Now()
		}
		query += fmt.Sprintf(" AND o.estimated_delivery_date IS NOT NULL AND date(o.estimated_delivery_date) < date($%d) AND COALESCE(p.total_paid, 0) < o.amount_charged", arg)
		args = append(args, today.Format("2006-01-02"))
		arg++
	}

	query += " ORDER BY datetime(o.entry_date) DESC, o.id DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []Order{}

	for rows.Next() {
		var o Order
		var status nullableOrderStatus
		var totalPaid float64
		if err := rows.Scan(
			&o.ID, &o.Description, &o.AmountCharged, &o.StatusID, &o.EntryDate,
			&o.EstimatedDeliveryDate, &o.DeliveryType, &o.Platform, &o.Notes,
			&o.ClientName, &o.ClientPhone,
			&o.PaidAt, &o.CreatedAt, &o.UpdatedAt,
			&status.ID, &status.Name, &status.DisplayName, &status.Color,
			&status.OrderPosition, &status.IsSystemStatus, &status.IsFinalStatus,
			&status.IsActive, &status.CreatedAt, &status.UpdatedAt,
			&totalPaid,
		); err != nil {
			return nil, err
		}

		o.Status = status.toOrderStatus()
		o.PaymentStatus = paymentStatusFromTotals(o.AmountCharged, totalPaid)

		orders = append(orders, o)
	}

	return orders, rows.Err()
}

type nullableOrderStatus struct {
	ID             sql.NullInt64
	Name           sql.NullString
	DisplayName    sql.NullString
	Color          sql.NullString
	OrderPosition  sql.NullInt64
	IsSystemStatus sql.NullInt64
	IsFinalStatus  sql.NullInt64
	IsActive       sql.NullInt64
	CreatedAt      internaldb.NullTime
	UpdatedAt      internaldb.NullTime
}

func (s nullableOrderStatus) toOrderStatus() *OrderStatus {
	if !s.ID.Valid {
		return nil
	}

	return &OrderStatus{
		ID:             int(s.ID.Int64),
		Name:           s.Name.String,
		DisplayName:    s.DisplayName.String,
		Color:          s.Color.String,
		OrderPosition:  int(s.OrderPosition.Int64),
		IsSystemStatus: s.IsSystemStatus.Int64 != 0,
		IsFinalStatus:  s.IsFinalStatus.Int64 != 0,
		IsActive:       s.IsActive.Int64 != 0,
		CreatedAt:      internaldb.Time{Time: s.CreatedAt.Time},
		UpdatedAt:      internaldb.Time{Time: s.UpdatedAt.Time},
	}
}

// -------------------- UPDATE --------------------

func (r *repository) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	query := `
	UPDATE orders SET
		description = COALESCE($1, description),
		amount_charged = COALESCE($2, amount_charged),
		status_id = COALESCE($3, status_id),
		estimated_delivery_date = COALESCE($4, estimated_delivery_date),
		delivery_type = COALESCE($5, delivery_type),
		platform = COALESCE($6, platform),
		notes = COALESCE($7, notes),
		client_name = COALESCE($8, client_name),
		client_phone = COALESCE($9, client_phone),
		updated_at = datetime('now')
	WHERE id = $10;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.Description,
		dto.AmountCharged,
		dto.StatusID,
		dto.EstimatedDeliveryDate,
		dto.DeliveryType,
		dto.Platform,
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
func (r *repository) Delete(ctx context.Context, id int, actorID *int) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET deleted_at = datetime('now'), deleted_by = $1, updated_at = datetime('now')
		WHERE id = $2 AND deleted_at IS NULL;
	`, actorID, id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order not found")
	}

	return nil
}

// -------------------- FINISH ORDER --------------------
func (r *repository) FinishOrder(ctx context.Context, id int) (*FinishOrderResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var amountCharged float64
	err = tx.QueryRowContext(ctx, "SELECT amount_charged FROM orders WHERE id = $1 AND deleted_at IS NULL", id).Scan(&amountCharged)
	if err == sql.ErrNoRows {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	var currentTotalPaid float64
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM income WHERE order_id = $1 AND deleted_at IS NULL", id).Scan(&currentTotalPaid)
	if err != nil {
		return nil, err
	}

	pendingAmount := amountCharged - currentTotalPaid
	amountPaid := 0.0
	var incomeID *int

	if pendingAmount > 0 {
		var createdIncomeID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO income (order_id, amount, date)
			VALUES ($1, $2, datetime('now'))
			RETURNING id;
		`, id, pendingAmount).Scan(&createdIncomeID)
		if err != nil {
			return nil, err
		}

		amountPaid = pendingAmount
		incomeID = &createdIncomeID
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE orders SET
			status_id = (SELECT id FROM order_statuses WHERE name = 'completed'),
			paid_at = datetime('now'),
			updated_at = datetime('now')
		WHERE id = $1 AND deleted_at IS NULL;
	`, id)
	if err != nil {
		return nil, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, errors.New("order not found")
	}

	var paidAt internaldb.NullTime
	if err := tx.QueryRowContext(ctx, "SELECT paid_at FROM orders WHERE id = $1", id).Scan(&paidAt); err != nil {
		return nil, err
	}

	totalPaid := currentTotalPaid + amountPaid
	remaining := amountCharged - totalPaid
	if remaining < 0 {
		remaining = 0
	}

	finishResult := &FinishOrderResult{
		Finished:      true,
		IncomeCreated: incomeID != nil,
		IncomeID:      incomeID,
		AmountPaid:    amountPaid,
		TotalPaid:     totalPaid,
		Remaining:     remaining,
		IsFullyPaid:   totalPaid >= amountCharged,
		PaidAt:        &paidAt,
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return finishResult, nil
}
