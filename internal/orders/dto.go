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
	Description           string       `json:"description"`
	AmountCharged         CustomFloat64 `json:"amount_charged"`
	Status                OrderStatus  `json:"status"`
	EstimatedDeliveryDate *time.Time   `json:"estimated_delivery_date"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	Notes                 *string      `json:"notes"`
}

type UpdateOrderDTO struct {
	Description           *string        `json:"description"`
	AmountCharged         *CustomFloat64 `json:"amount_charged"`
	Status                *OrderStatus   `json:"status"`
	EstimatedDeliveryDate *time.Time     `json:"estimated_delivery_date"`
	DeliveryType          *DeliveryType  `json:"delivery_type"`
	Notes                 *string        `json:"notes"`
}
