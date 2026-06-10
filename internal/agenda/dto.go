package agenda

import "time"

type CreateAgendaItemDTO struct {
	Type     ItemType     `json:"type"`
	Title    string       `json:"title"`
	Content  *string      `json:"content"`
	Priority ItemPriority `json:"priority"`
	DueDate  *time.Time   `json:"due_date"`
	OrderID  *int         `json:"order_id"`
}

type UpdateAgendaItemDTO struct {
	Type     *ItemType     `json:"type"`
	Title    *string       `json:"title"`
	Content  *string       `json:"content"`
	Status   *ItemStatus   `json:"status"`
	Priority *ItemPriority `json:"priority"`
	DueDate  *time.Time    `json:"due_date"`
	OrderID  *int          `json:"order_id"`
}

type FilterAgendaItemsDTO struct {
	// "pending" | "done" | "archived" | "all" — defaults to "pending"
	Status   string
	Type     string
	Priority string
	OrderID  *int
	From     *string
	To       *string
	Search   *string
}
