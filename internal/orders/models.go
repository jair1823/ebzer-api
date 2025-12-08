package orders

import "time"

type OrderStatus string
type DeliveryType string

const (
	StatusPending   OrderStatus = "pending"
	StatusCompleted OrderStatus = "completed"
	StatusPaid      OrderStatus = "paid"
)

const (
	DeliveryPickup   DeliveryType = "pickup"
	DeliveryShipping DeliveryType = "shipping"
	DeliveryDelivery DeliveryType = "delivery"
)

type Order struct {
	ID                    int          `json:"id"`
	Description           string       `json:"description"`
	AmountCharged         float64      `json:"amount_charged"`
	Status                OrderStatus  `json:"status"`
	EntryDate             time.Time    `json:"entry_date"`
	EstimatedDeliveryDate *time.Time   `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	ClientName            *string      `json:"client_name"`
	ClientPhone           *string      `json:"client_phone"`
	Notes                 *string      `json:"notes"`
	Paid50Percent         bool         `json:"paid_50_percent"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}
