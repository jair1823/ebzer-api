package expenses

import "creaciones-api/internal/db"

type Comercio struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   db.Time `json:"created_at"`
}

type Product struct {
	ID           int       `json:"id"`
	ComercioID   int       `json:"comercio_id"`
	Comercio     *Comercio `json:"comercio,omitempty"`
	Name         string    `json:"name"`
	DefaultPrice float64   `json:"default_price"`
	CreatedAt    db.Time   `json:"created_at"`
}

type ExpenseItem struct {
	ID          int     `json:"id"`
	ExpenseID   int     `json:"expense_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
}

type Expense struct {
	ID          int           `json:"id"`
	ComercioID  int           `json:"comercio_id"`
	Comercio    *Comercio     `json:"comercio,omitempty"`
	Description *string       `json:"description"`
	Date        db.Time       `json:"date"`
	Amount      *float64      `json:"amount"`
	Total       float64       `json:"total"`
	Items       []ExpenseItem `json:"items"`
	CreatedAt   db.Time       `json:"created_at"`
}
