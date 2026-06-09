package orders

import (
	"context"
	"database/sql"
	"errors"
)

// StatusRepository handles persistence for order_statuses table
type StatusRepository interface {
	Create(ctx context.Context, dto CreateOrderStatusDTO) (int, error)
	GetByID(ctx context.Context, id int) (*OrderStatus, error)
	GetByName(ctx context.Context, name string) (*OrderStatus, error)
	GetAll(ctx context.Context, activeOnly bool) ([]OrderStatus, error)
	Update(ctx context.Context, id int, dto UpdateOrderStatusDTO) error
	Deactivate(ctx context.Context, id int) error
	Reorder(ctx context.Context, items []StatusOrderItem) error
	CountOrdersForStatus(ctx context.Context, statusID int) (int, error)
}

type statusRepository struct {
	db *sql.DB
}

func NewStatusRepository(db *sql.DB) StatusRepository {
	return &statusRepository{db: db}
}

const statusScanCols = `id, name, display_name, color, order_position,
	is_system_status, is_final_status, is_active, created_at, updated_at`

func scanStatus(row interface {
	Scan(dest ...any) error
}) (*OrderStatus, error) {
	var s OrderStatus
	var isSystem, isFinal, isActive int
	err := row.Scan(
		&s.ID, &s.Name, &s.DisplayName, &s.Color, &s.OrderPosition,
		&isSystem, &isFinal, &isActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.IsSystemStatus = isSystem != 0
	s.IsFinalStatus = isFinal != 0
	s.IsActive = isActive != 0
	return &s, nil
}

// -------------------- CREATE --------------------

func (r *statusRepository) Create(ctx context.Context, dto CreateOrderStatusDTO) (int, error) {
	query := `
	INSERT INTO order_statuses (name, display_name, color, order_position, is_final_status)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`
	color := dto.Color
	if color == "" {
		color = "#6B7280"
	}

	var id int
	err := r.db.QueryRowContext(ctx, query,
		dto.Name,
		dto.DisplayName,
		color,
		dto.OrderPosition,
		boolToInt(dto.IsFinalStatus),
	).Scan(&id)
	return id, err
}

// -------------------- GET BY ID --------------------

func (r *statusRepository) GetByID(ctx context.Context, id int) (*OrderStatus, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+statusScanCols+" FROM order_statuses WHERE id = $1", id)
	s, err := scanStatus(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

// -------------------- GET BY NAME --------------------

func (r *statusRepository) GetByName(ctx context.Context, name string) (*OrderStatus, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+statusScanCols+" FROM order_statuses WHERE LOWER(name) = LOWER($1)", name)
	s, err := scanStatus(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

// -------------------- GET ALL --------------------

func (r *statusRepository) GetAll(ctx context.Context, activeOnly bool) ([]OrderStatus, error) {
	query := "SELECT " + statusScanCols + " FROM order_statuses"
	if activeOnly {
		query += " WHERE is_active = 1"
	}
	query += " ORDER BY order_position ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []OrderStatus
	for rows.Next() {
		s, err := scanStatus(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *s)
	}
	return list, rows.Err()
}

// -------------------- UPDATE --------------------

func (r *statusRepository) Update(ctx context.Context, id int, dto UpdateOrderStatusDTO) error {
	query := `
	UPDATE order_statuses SET
		display_name   = COALESCE($1, display_name),
		color          = COALESCE($2, color),
		order_position = COALESCE($3, order_position),
		is_final_status = COALESCE($4, is_final_status),
		updated_at     = datetime('now')
	WHERE id = $5;
	`
	var isFinal *int
	if dto.IsFinalStatus != nil {
		v := boolToInt(*dto.IsFinalStatus)
		isFinal = &v
	}

	result, err := r.db.ExecContext(ctx, query,
		dto.DisplayName,
		dto.Color,
		dto.OrderPosition,
		isFinal,
		id,
	)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order status not found")
	}
	return nil
}

// -------------------- DEACTIVATE (soft delete) --------------------

func (r *statusRepository) Deactivate(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE order_statuses SET is_active = 0, updated_at = datetime('now') WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("order status not found")
	}
	return nil
}

// -------------------- REORDER --------------------

func (r *statusRepository) Reorder(ctx context.Context, items []StatusOrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE order_statuses SET order_position = $1, updated_at = datetime('now') WHERE id = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, item.Position, item.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// -------------------- COUNT ORDERS FOR STATUS --------------------

func (r *statusRepository) CountOrdersForStatus(ctx context.Context, statusID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM orders WHERE status_id = $1", statusID).Scan(&count)
	return count, err
}

// -------------------- HELPERS --------------------

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
