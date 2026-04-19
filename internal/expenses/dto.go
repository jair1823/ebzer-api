package expenses

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

// Expense DTOs
type CreateExpenseDTO struct {
	Description string        `json:"description"`
	Amount      CustomFloat64 `json:"amount"`
	Date        *string       `json:"date"`
	OrderID     *int          `json:"order_id"`
	CategoryID  *int          `json:"category_id"`
	Type        string        `json:"type"`
}

type UpdateExpenseDTO struct {
	Description *string        `json:"description"`
	Amount      *CustomFloat64 `json:"amount"`
	Date        *string        `json:"date"`
	OrderID     *int           `json:"order_id"`
	CategoryID  *int           `json:"category_id"`
	Type        *string        `json:"type"`
}

// Category DTOs
type CreateCategoryDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateCategoryDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
