package orders

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StatusHandler struct {
	service StatusService
}

func NewStatusHandler(s StatusService) *StatusHandler {
	return &StatusHandler{service: s}
}

func (h *StatusHandler) RegisterRoutes(router fiber.Router) {
	// PUT /reorder must be registered before /:id to avoid being caught by param route
	router.Put("/reorder", h.Reorder)
	router.Get("/", h.GetAll)
	router.Post("/", h.Create)
	router.Get("/:id", h.GetByID)
	router.Put("/:id", h.Update)
	router.Delete("/:id", h.Deactivate)
}

// -------------------- GET ALL --------------------

func (h *StatusHandler) GetAll(c *fiber.Ctx) error {
	activeOnly := c.Query("active_only") == "true"
	list, err := h.service.GetAll(c.Context(), activeOnly)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"statuses": list})
}

// -------------------- GET BY ID --------------------

func (h *StatusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	s, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s == nil {
		return fiber.NewError(fiber.StatusNotFound, "order status not found")
	}
	return c.JSON(s)
}

// -------------------- CREATE --------------------

func (h *StatusHandler) Create(c *fiber.Ctx) error {
	var dto CreateOrderStatusDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	id, err := h.service.Create(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// -------------------- UPDATE --------------------

func (h *StatusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var dto UpdateOrderStatusDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.service.Update(c.Context(), id, dto); err != nil {
		if err.Error() == "order status not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"updated": true})
}

// -------------------- DEACTIVATE --------------------

func (h *StatusHandler) Deactivate(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.service.Deactivate(c.Context(), id); err != nil {
		if err.Error() == "order status not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"deactivated": true})
}

// -------------------- REORDER --------------------

func (h *StatusHandler) Reorder(c *fiber.Ctx) error {
	var dto ReorderStatusesDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.service.Reorder(c.Context(), dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"reordered": true})
}
