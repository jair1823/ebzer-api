package orders

import (
	"creaciones-api/internal/db"
)

type DeliveryType string

const (
	DeliveryPickup   DeliveryType = "pickup"
	DeliveryShipping DeliveryType = "shipping"
	DeliveryDelivery DeliveryType = "delivery"
)

// OrderStatus represents a configurable status stored in order_statuses table
type OrderStatus struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	DisplayName    string  `json:"display_name"`
	Color          string  `json:"color"`
	OrderPosition  int     `json:"order_position"`
	IsSystemStatus bool    `json:"is_system_status"`
	IsFinalStatus  bool    `json:"is_final_status"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      db.Time `json:"created_at"`
	UpdatedAt      db.Time `json:"updated_at"`
}

type Order struct {
	ID                    int          `json:"id"`
	Description           string       `json:"description"`
	AmountCharged         float64      `json:"amount_charged"`
	StatusID              int          `json:"status_id"`
	Status                *OrderStatus `json:"status,omitempty"`
	EntryDate             db.Time      `json:"entry_date"`
	EstimatedDeliveryDate *db.NullTime `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	ClientName            *string      `json:"client_name"`
	ClientPhone           *string      `json:"client_phone"`
	Notes                 *string      `json:"notes"`
	PaidAt                *db.NullTime `json:"paid_at"`
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

type FinishOrderResult struct {
	Finished      bool         `json:"finished"`
	IncomeCreated bool         `json:"income_created"`
	IncomeID      *int         `json:"income_id"`
	AmountPaid    float64      `json:"amount_paid"`
	TotalPaid     float64      `json:"total_paid"`
	Remaining     float64      `json:"remaining"`
	IsFullyPaid   bool         `json:"is_fully_paid"`
	PaidAt        *db.NullTime `json:"paid_at"`
}
