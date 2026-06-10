package agenda

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	internaldb "creaciones-api/internal/db"
)

type Repository interface {
	Create(ctx context.Context, dto CreateAgendaItemDTO) (int, error)
	GetByID(ctx context.Context, id int) (*AgendaItem, error)
	GetAll(ctx context.Context, filter FilterAgendaItemsDTO) ([]AgendaItem, error)
	Update(ctx context.Context, id int, dto UpdateAgendaItemDTO) error
	Delete(ctx context.Context, id int) error
	Complete(ctx context.Context, id int) error
	Archive(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// -------------------- CREATE --------------------

func (r *repository) Create(ctx context.Context, dto CreateAgendaItemDTO) (int, error) {
	// Apply defaults so the repo can be called directly without going through the service.
	if dto.Type == "" {
		dto.Type = TypeNote
	}
	if dto.Priority == "" {
		dto.Priority = PriorityMedium
	}

	query := `
	INSERT INTO agenda_items (type, title, content, priority, due_date, order_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Type,
		dto.Title,
		dto.Content,
		dto.Priority,
		dto.DueDate,
		dto.OrderID,
	).Scan(&id)

	return id, err
}

// -------------------- GET BY ID --------------------

func (r *repository) GetByID(ctx context.Context, id int) (*AgendaItem, error) {
	row := r.db.QueryRowContext(ctx, `
	SELECT
		a.id, a.type, a.title, a.content, a.status, a.priority,
		a.due_date, a.completed_at, a.order_id,
		a.created_at, a.updated_at,
		o.id, o.description, o.client_name, o.client_phone,
		o.estimated_delivery_date, o.status_id
	FROM agenda_items a
	LEFT JOIN orders o ON o.id = a.order_id
	WHERE a.id = $1;
	`, id)

	return scanAgendaItem(row)
}

// -------------------- GET ALL (FILTERS) --------------------

func (r *repository) GetAll(ctx context.Context, filter FilterAgendaItemsDTO) ([]AgendaItem, error) {
	query := `
	SELECT
		a.id, a.type, a.title, a.content, a.status, a.priority,
		a.due_date, a.completed_at, a.order_id,
		a.created_at, a.updated_at,
		o.id, o.description, o.client_name, o.client_phone,
		o.estimated_delivery_date, o.status_id
	FROM agenda_items a
	LEFT JOIN orders o ON o.id = a.order_id
	WHERE 1 = 1
	`

	args := []any{}
	arg := 1

	// status filter: default "pending", "all" skips filter
	if filter.Status != "" && filter.Status != "all" {
		query += fmt.Sprintf(" AND a.status = $%d", arg)
		args = append(args, filter.Status)
		arg++
	}

	if filter.Type != "" {
		query += fmt.Sprintf(" AND a.type = $%d", arg)
		args = append(args, filter.Type)
		arg++
	}

	if filter.Priority != "" {
		query += fmt.Sprintf(" AND a.priority = $%d", arg)
		args = append(args, filter.Priority)
		arg++
	}

	if filter.OrderID != nil {
		query += fmt.Sprintf(" AND a.order_id = $%d", arg)
		args = append(args, *filter.OrderID)
		arg++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND a.due_date >= $%d", arg)
		args = append(args, *filter.From)
		arg++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND a.due_date <= $%d", arg)
		args = append(args, *filter.To)
		arg++
	}

	if filter.Search != nil && *filter.Search != "" {
		term := "%" + *filter.Search + "%"
		query += fmt.Sprintf(
			" AND (a.title LIKE $%d OR a.content LIKE $%d OR o.description LIKE $%d OR o.client_name LIKE $%d)",
			arg, arg+1, arg+2, arg+3,
		)
		args = append(args, term, term, term, term)
		arg += 4
	}

	query += " ORDER BY a.created_at DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AgendaItem{}
	for rows.Next() {
		item, err := scanAgendaItemRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

// -------------------- UPDATE --------------------

func (r *repository) Update(ctx context.Context, id int, dto UpdateAgendaItemDTO) error {
	query := `
	UPDATE agenda_items SET
		type        = COALESCE($1, type),
		title       = COALESCE($2, title),
		content     = COALESCE($3, content),
		status      = COALESCE($4, status),
		priority    = COALESCE($5, priority),
		due_date    = COALESCE($6, due_date),
		order_id    = COALESCE($7, order_id),
		updated_at  = datetime('now')
	WHERE id = $8;
	`

	result, err := r.db.ExecContext(ctx, query,
		dto.Type,
		dto.Title,
		dto.Content,
		dto.Status,
		dto.Priority,
		dto.DueDate,
		dto.OrderID,
		id,
	)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("agenda item not found")
	}

	return nil
}

// -------------------- DELETE --------------------

func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM agenda_items WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("agenda item not found")
	}

	return nil
}

// -------------------- COMPLETE --------------------

func (r *repository) Complete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `
	UPDATE agenda_items SET
		status       = 'done',
		completed_at = datetime('now'),
		updated_at   = datetime('now')
	WHERE id = $1;
	`, id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("agenda item not found")
	}

	return nil
}

// -------------------- ARCHIVE --------------------

func (r *repository) Archive(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `
	UPDATE agenda_items SET
		status     = 'archived',
		updated_at = datetime('now')
	WHERE id = $1;
	`, id)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("agenda item not found")
	}

	return nil
}

// -------------------- SCAN HELPERS --------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAgendaItem(scanner rowScanner) (*AgendaItem, error) {
	var item AgendaItem
	var ns nullableOrderSummary

	err := scanner.Scan(
		&item.ID, &item.Type, &item.Title, &item.Content,
		&item.Status, &item.Priority,
		&item.DueDate, &item.CompletedAt,
		&item.OrderID,
		&item.CreatedAt, &item.UpdatedAt,
		&ns.ID, &ns.Description, &ns.ClientName, &ns.ClientPhone,
		&ns.EstimatedDeliveryDate, &ns.StatusID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	item.Order = ns.toOrderSummary()
	return &item, nil
}

func scanAgendaItemRow(rows *sql.Rows) (*AgendaItem, error) {
	var item AgendaItem
	var ns nullableOrderSummary

	err := rows.Scan(
		&item.ID, &item.Type, &item.Title, &item.Content,
		&item.Status, &item.Priority,
		&item.DueDate, &item.CompletedAt,
		&item.OrderID,
		&item.CreatedAt, &item.UpdatedAt,
		&ns.ID, &ns.Description, &ns.ClientName, &ns.ClientPhone,
		&ns.EstimatedDeliveryDate, &ns.StatusID,
	)
	if err != nil {
		return nil, err
	}

	item.Order = ns.toOrderSummary()
	return &item, nil
}

type nullableOrderSummary struct {
	ID                    sql.NullInt64
	Description           sql.NullString
	ClientName            sql.NullString
	ClientPhone           sql.NullString
	EstimatedDeliveryDate internaldb.NullTime
	StatusID              sql.NullInt64
}

func (ns nullableOrderSummary) toOrderSummary() *OrderSummary {
	if !ns.ID.Valid {
		return nil
	}

	var clientName *string
	if ns.ClientName.Valid {
		clientName = &ns.ClientName.String
	}

	var clientPhone *string
	if ns.ClientPhone.Valid {
		clientPhone = &ns.ClientPhone.String
	}

	var edd *internaldb.NullTime
	if ns.EstimatedDeliveryDate.Valid {
		edd = &ns.EstimatedDeliveryDate
	}

	return &OrderSummary{
		ID:                    int(ns.ID.Int64),
		Description:           ns.Description.String,
		ClientName:            clientName,
		ClientPhone:           clientPhone,
		EstimatedDeliveryDate: edd,
		StatusID:              int(ns.StatusID.Int64),
	}
}
