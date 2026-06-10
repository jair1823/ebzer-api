package agenda

import "creaciones-api/internal/db"

type ItemType string
type ItemStatus string
type ItemPriority string

const (
	TypeNote     ItemType = "note"
	TypeTask     ItemType = "task"
	TypeReminder ItemType = "reminder"
)

const (
	StatusPending  ItemStatus = "pending"
	StatusDone     ItemStatus = "done"
	StatusArchived ItemStatus = "archived"
)

const (
	PriorityLow    ItemPriority = "low"
	PriorityMedium ItemPriority = "medium"
	PriorityHigh   ItemPriority = "high"
)

// OrderSummary is the linked order data included on each agenda item when order_id is set.
type OrderSummary struct {
	ID                    int          `json:"id"`
	Description           string       `json:"description"`
	ClientName            *string      `json:"client_name"`
	ClientPhone           *string      `json:"client_phone"`
	EstimatedDeliveryDate *db.NullTime `json:"estimated_delivery_date"`
	StatusID              int          `json:"status_id"`
}

type AgendaItem struct {
	ID          int          `json:"id"`
	Type        ItemType     `json:"type"`
	Title       string       `json:"title"`
	Content     *string      `json:"content"`
	Status      ItemStatus   `json:"status"`
	Priority    ItemPriority `json:"priority"`
	DueDate     *db.NullTime `json:"due_date"`
	CompletedAt *db.NullTime `json:"completed_at"`
	OrderID     *int         `json:"order_id"`
	Order       *OrderSummary `json:"order,omitempty"`
	CreatedAt   db.Time      `json:"created_at"`
	UpdatedAt   db.Time      `json:"updated_at"`
}
