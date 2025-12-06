package orders

import "time"

type OrderStatus string
type DeliveryType string

const (
	StatusNew            OrderStatus = "new"
	StatusDesign         OrderStatus = "design"
	StatusPendingClient  OrderStatus = "pending_client"
	StatusProduction     OrderStatus = "production"
	StatusReady          OrderStatus = "ready"
	StatusDelivered      OrderStatus = "delivered"
)

const (
	DeliveryPickup   DeliveryType = "pickup"
	DeliveryShipping DeliveryType = "shipping"
	DeliveryDelivery DeliveryType = "delivery"
	DeliveryOther    DeliveryType = "other"
)

type Order struct {
	ID                    int          `json:"id"`
	Description           string       `json:"description"`
	AmountCharged         float64      `json:"amount_charged"`
	Status                OrderStatus  `json:"status"`
	EntryDate             time.Time    `json:"entry_date"`
	EstimatedDeliveryDate *time.Time   `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	Notes                 *string      `json:"notes"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}
