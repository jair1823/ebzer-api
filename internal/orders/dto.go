package orders

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// CustomFloat64 permite unmarshal de string o número
type CustomFloat64 float64

func (c *CustomFloat64) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*c = CustomFloat64(value)
	case string:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string to float64: %v", err)
		}
		*c = CustomFloat64(f)
	default:
		return fmt.Errorf("amount_charged must be a number or string")
	}
	return nil
}

type CreateOrderDTO struct {
	Description           string        `json:"description"`
	AmountCharged         CustomFloat64 `json:"amount_charged"`
	StatusID              int           `json:"status_id"`
	EstimatedDeliveryDate *time.Time    `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType  `json:"delivery_type"`
	ClientName            *string       `json:"client_name"`
	ClientPhone           *string       `json:"client_phone"`
	Notes                 *string       `json:"notes"`
}

type UpdateOrderDTO struct {
	Description           *string        `json:"description"`
	AmountCharged         *CustomFloat64 `json:"amount_charged"`
	StatusID              *int           `json:"status_id"`
	EstimatedDeliveryDate *time.Time     `json:"estimated_delivery_date"`
	DeliveryType          *DeliveryType  `json:"delivery_type"`
	ClientName            *string        `json:"client_name"`
	ClientPhone           *string        `json:"client_phone"`
	Notes                 *string        `json:"notes"`
}

// ---- Order Status DTOs ----

type CreateOrderStatusDTO struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Color         string `json:"color"`
	OrderPosition int    `json:"order_position"`
	IsFinalStatus bool   `json:"is_final_status"`
}

type UpdateOrderStatusDTO struct {
	DisplayName   *string `json:"display_name"`
	Color         *string `json:"color"`
	OrderPosition *int    `json:"order_position"`
	IsFinalStatus *bool   `json:"is_final_status"`
	IsActive      *bool   `json:"is_active"`
}

type ReorderStatusesDTO struct {
	StatusOrders []StatusOrderItem `json:"status_orders"`
}

type StatusOrderItem struct {
	ID       int `json:"id"`
	Position int `json:"position"`
}
