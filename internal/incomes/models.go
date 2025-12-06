package incomes

import "time"

type Income struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
