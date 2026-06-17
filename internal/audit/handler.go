package audit

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	filter := FilterDTO{
		EntityType: queryStringPtr(c, "entity_type"),
		From:       queryStringPtr(c, "from"),
		To:         queryStringPtr(c, "to"),
	}

	if entityID := c.Query("entity_id"); entityID != "" {
		id, err := strconv.Atoi(entityID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid entity_id")
		}
		filter.EntityID = &id
	}

	events, err := h.repo.GetAll(c.Context(), filter)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(events)
}

func queryStringPtr(c *fiber.Ctx, name string) *string {
	value := c.Query(name)
	if value == "" {
		return nil
	}
	return &value
}
