package expenses

import (
	"creaciones-api/internal/db"
)

type ExpenseType string

const (
	TypeGeneral     ExpenseType = "general"
	TypeOrderLinked ExpenseType = "order_linked"
)

type Expense struct {
	ID          int         `json:"id"`
	Description string      `json:"description"`
	Amount      float64     `json:"amount"`
	Date        db.Time     `json:"date"`
	OrderID     *int        `json:"order_id"`
	CategoryID  *int        `json:"category_id"`
	Type        ExpenseType `json:"type"`
	CreatedAt   db.Time     `json:"created_at"`
}

type ExpenseCategory struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   db.Time `json:"created_at"`
}
