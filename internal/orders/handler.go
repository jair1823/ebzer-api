package orders

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/", h.Create)
	router.Get("/", h.GetAll)
	router.Get("/:id", h.GetByID)
	router.Put("/:id", h.Update)
	router.Post("/:id/finish", h.FinishOrder)
	router.Delete("/:id", h.Delete)
}

// -------------------- CREATE --------------------

func (h *Handler) Create(c *fiber.Ctx) error {
	var dto CreateOrderDTO

	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := h.service.Create(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// -------------------- GET ALL --------------------

func (h *Handler) GetAll(c *fiber.Ctx) error {
	statusStr := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	var status *OrderStatus
	if statusStr != "" {
		s := OrderStatus(statusStr)
		status = &s
	}

	// Solo pasar punteros si los valores no están vacíos
	var fromPtr *string
	if from != "" {
		fromPtr = &from
	}

	var toPtr *string
	if to != "" {
		toPtr = &to
	}

	orders, err := h.service.GetAll(c.Context(), status, fromPtr, toPtr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(orders)
}

// -------------------- GET BY ID --------------------

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	order, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if order == nil {
		return fiber.NewError(404, "order not found")
	}

	return c.JSON(order)
}

// -------------------- UPDATE --------------------

func (h *Handler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateOrderDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.Update(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "order not found" {
			return fiber.NewError(404, err.Error())
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

// -------------------- DELETE --------------------

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.Delete(c.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			return fiber.NewError(404, "order not found")
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// -------------------- FINISH ORDER --------------------
func (h *Handler) FinishOrder(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	err := h.service.FinishOrder(c.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			return fiber.NewError(404, "order not found")
		}
		return fiber.NewError(500, err.Error())
	}
	return c.JSON(fiber.Map{"finished": true})
}
