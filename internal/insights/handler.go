package insights

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	summary, err := h.service.GetSummary(c.Context(), queryPtr(c, "from"), queryPtr(c, "to"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(summary)
}

func queryPtr(c *fiber.Ctx, name string) *string {
	value := c.Query(name)
	if value == "" {
		return nil
	}
	return &value
}
