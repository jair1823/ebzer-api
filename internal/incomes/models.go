package incomes

import (
	"creaciones-api/internal/db"
)

type Income struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	Amount    float64 `json:"amount"`
	Date      db.Time `json:"date"`
	CreatedAt db.Time `json:"created_at"`
	UpdatedAt db.Time `json:"updated_at"`
}
