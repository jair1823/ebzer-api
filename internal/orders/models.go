package orders

import (
	"creaciones-api/internal/db"
)

type OrderStatus string
type DeliveryType string

const (
	StatusConfirmed  OrderStatus = "confirmed"
	StatusInProgress OrderStatus = "in_progress"
	StatusReady      OrderStatus = "ready"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
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
	EntryDate             db.Time      `json:"entry_date"`
	EstimatedDeliveryDate *db.NullTime `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	ClientName            *string      `json:"client_name"`
	ClientPhone           *string      `json:"client_phone"`
	Notes                 *string      `json:"notes"`
	CreatedAt             db.Time      `json:"created_at"`
	UpdatedAt             db.Time      `json:"updated_at"`
}

type PaymentStatus struct {
	TotalPaid      float64 `json:"total_paid"`
	AmountCharged  float64 `json:"amount_charged"`
	Remaining      float64 `json:"remaining"`
	PercentagePaid float64 `json:"percentage_paid"`
	IsFullyPaid    bool    `json:"is_fully_paid"`
}
