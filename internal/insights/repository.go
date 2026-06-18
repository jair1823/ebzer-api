package insights

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetSummary(ctx context.Context, filter SummaryFilter) (*Summary, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetSummary(ctx context.Context, filter SummaryFilter) (*Summary, error) {
	summary := &Summary{
		SalesByPlatform:     []PlatformSales{},
		TopExpenseMerchants: []MerchantExpense{},
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM income
		WHERE deleted_at IS NULL
			AND date(date) >= date($1)
			AND date(date) <= date($2);
	`, filter.From, filter.To).Scan(&summary.IncomeTotal); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		WITH expense_totals AS (
			SELECT e.id, COALESCE(e.amount, SUM(ei.quantity * ei.unit_price), 0) AS total
			FROM expenses e
			LEFT JOIN expense_items ei ON ei.expense_id = e.id
			WHERE e.deleted_at IS NULL
				AND date(e.date) >= date($1)
				AND date(e.date) <= date($2)
			GROUP BY e.id
		)
		SELECT COALESCE(SUM(total), 0)
		FROM expense_totals;
	`, filter.From, filter.To).Scan(&summary.ExpenseTotal); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		WITH paid AS (
			SELECT order_id, COALESCE(SUM(amount), 0) AS total_paid
			FROM income
			WHERE deleted_at IS NULL
			GROUP BY order_id
		)
		SELECT COALESCE(SUM(MAX(o.amount_charged - COALESCE(p.total_paid, 0), 0)), 0)
		FROM orders o
		LEFT JOIN paid p ON p.order_id = o.id
		WHERE o.deleted_at IS NULL
			AND COALESCE(p.total_paid, 0) < o.amount_charged;
	`).Scan(&summary.PendingCollection); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM orders o
		LEFT JOIN order_statuses s ON s.id = o.status_id
		WHERE o.deleted_at IS NULL
			AND date(o.entry_date) >= date($1)
			AND date(o.entry_date) <= date($2)
			AND COALESCE(s.is_final_status, 0) = 0;
	`, filter.From, filter.To).Scan(&summary.ActiveOrders); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		WITH paid AS (
			SELECT order_id, COALESCE(SUM(amount), 0) AS total_paid
			FROM income
			WHERE deleted_at IS NULL
			GROUP BY order_id
		)
		SELECT COUNT(*)
		FROM orders o
		LEFT JOIN order_statuses s ON s.id = o.status_id
		LEFT JOIN paid p ON p.order_id = o.id
		WHERE o.deleted_at IS NULL
			AND date(o.entry_date) >= date($1)
			AND date(o.entry_date) <= date($2)
			AND COALESCE(s.is_final_status, 0) = 1
			AND COALESCE(p.total_paid, 0) >= o.amount_charged
			AND COALESCE(p.total_paid, 0) > 0;
	`, filter.From, filter.To).Scan(&summary.PaidCompletedOrders); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		WITH paid AS (
			SELECT order_id, COALESCE(SUM(amount), 0) AS total_paid
			FROM income
			WHERE deleted_at IS NULL
			GROUP BY order_id
		)
		SELECT COUNT(*)
		FROM orders o
		LEFT JOIN paid p ON p.order_id = o.id
		WHERE o.deleted_at IS NULL
			AND o.estimated_delivery_date IS NOT NULL
			AND date(o.estimated_delivery_date) < date('now')
			AND COALESCE(p.total_paid, 0) < o.amount_charged;
	`).Scan(&summary.OverdueOrders); err != nil {
		return nil, err
	}

	salesRows, err := r.db.QueryContext(ctx, `
		SELECT platform, COUNT(*), COALESCE(SUM(amount_charged), 0)
		FROM orders
		WHERE deleted_at IS NULL
			AND date(entry_date) >= date($1)
			AND date(entry_date) <= date($2)
		GROUP BY platform
		ORDER BY SUM(amount_charged) DESC;
	`, filter.From, filter.To)
	if err != nil {
		return nil, err
	}
	defer salesRows.Close()
	for salesRows.Next() {
		var item PlatformSales
		if err := salesRows.Scan(&item.Platform, &item.Count, &item.Total); err != nil {
			return nil, err
		}
		summary.SalesByPlatform = append(summary.SalesByPlatform, item)
	}
	if err := salesRows.Err(); err != nil {
		return nil, err
	}

	merchantRows, err := r.db.QueryContext(ctx, `
		WITH expense_totals AS (
			SELECT
				e.id,
				e.comercio_id,
				COALESCE(e.amount, SUM(ei.quantity * ei.unit_price), 0) AS total
			FROM expenses e
			LEFT JOIN expense_items ei ON ei.expense_id = e.id
			WHERE e.deleted_at IS NULL
				AND date(e.date) >= date($1)
				AND date(e.date) <= date($2)
			GROUP BY e.id
		)
		SELECT c.id, c.name, COALESCE(SUM(et.total), 0) AS total
		FROM expense_totals et
		INNER JOIN comercios c ON c.id = et.comercio_id
		GROUP BY c.id, c.name
		ORDER BY total DESC
		LIMIT 5;
	`, filter.From, filter.To)
	if err != nil {
		return nil, err
	}
	defer merchantRows.Close()
	for merchantRows.Next() {
		var item MerchantExpense
		if err := merchantRows.Scan(&item.ComercioID, &item.Name, &item.Total); err != nil {
			return nil, err
		}
		summary.TopExpenseMerchants = append(summary.TopExpenseMerchants, item)
	}
	if err := merchantRows.Err(); err != nil {
		return nil, err
	}

	summary.Profit = summary.IncomeTotal - summary.ExpenseTotal
	return summary, nil
}
