package expenses

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

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
		return fmt.Errorf("value must be a number or string")
	}
	return nil
}

type CustomInt int

func (c *CustomInt) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		if value != math.Trunc(value) {
			return fmt.Errorf("value must be an integer")
		}
		*c = CustomInt(int(value))
	case string:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string to integer: %v", err)
		}
		if f != math.Trunc(f) {
			return fmt.Errorf("value must be an integer")
		}
		*c = CustomInt(int(f))
	default:
		return fmt.Errorf("value must be a number or string")
	}
	return nil
}

type CreateExpenseItemDTO struct {
	ProductID   *int          `json:"product_id"`
	ProductName string        `json:"product_name"`
	Quantity    CustomInt     `json:"quantity"`
	UnitPrice   CustomFloat64 `json:"unit_price"`
}

type CreateExpenseDTO struct {
	ComercioID  int                    `json:"comercio_id"`
	Date        *string                `json:"date"`
	Description *string                `json:"description"`
	Amount      *CustomFloat64         `json:"amount"`
	Items       []CreateExpenseItemDTO `json:"items"`
}

type UpdateExpenseDTO struct {
	ComercioID  int                    `json:"comercio_id"`
	Date        *string                `json:"date"`
	Description *string                `json:"description"`
	Amount      *CustomFloat64         `json:"amount"`
	Items       []CreateExpenseItemDTO `json:"items"`
}

type CreateComercioDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateComercioDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateProductDTO struct {
	ComercioID   int           `json:"comercio_id"`
	Name         string        `json:"name"`
	DefaultPrice CustomFloat64 `json:"default_price"`
}

type UpdateProductDTO struct {
	ComercioID   *int           `json:"comercio_id"`
	Name         *string        `json:"name"`
	DefaultPrice *CustomFloat64 `json:"default_price"`
}
