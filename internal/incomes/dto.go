package incomes

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// CustomFloat64 allows unmarshalling from string or number
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
		return fmt.Errorf("amount must be a number or string")
	}
	return nil
}

type CreateIncomeDTO struct {
	OrderID string        `json:"order_id"`
	Amount  CustomFloat64 `json:"amount"`
}

type UpdateIncomeDTO struct {
	OrderID *string        `json:"order_id"`
	Amount  *CustomFloat64 `json:"amount"`
}
