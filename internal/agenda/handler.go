package agenda

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
	router.Patch("/:id", h.Update)
	router.Delete("/:id", h.Delete)
	router.Patch("/:id/complete", h.Complete)
	router.Patch("/:id/archive", h.Archive)
}

// -------------------- CREATE --------------------

func (h *Handler) Create(c *fiber.Ctx) error {
	var dto CreateAgendaItemDTO
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
	filter := FilterAgendaItemsDTO{
		Status:   c.Query("status"),
		Type:     c.Query("type"),
		Priority: c.Query("priority"),
	}

	if orderIDStr := c.Query("order_id"); orderIDStr != "" {
		if id, err := strconv.Atoi(orderIDStr); err == nil {
			filter.OrderID = &id
		}
	}

	if from := c.Query("from"); from != "" {
		filter.From = &from
	}

	if to := c.Query("to"); to != "" {
		filter.To = &to
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	items, err := h.service.GetAll(c.Context(), filter)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(items)
}

// -------------------- GET BY ID --------------------

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if item == nil {
		return fiber.NewError(fiber.StatusNotFound, "agenda item not found")
	}

	return c.JSON(item)
}

// -------------------- UPDATE --------------------

func (h *Handler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateAgendaItemDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.Update(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "agenda item not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

// -------------------- DELETE --------------------

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.Delete(c.Context(), id)
	if err != nil {
		if err.Error() == "agenda item not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// -------------------- COMPLETE --------------------

func (h *Handler) Complete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.Complete(c.Context(), id)
	if err != nil {
		if err.Error() == "agenda item not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"completed": true})
}

// -------------------- ARCHIVE --------------------

func (h *Handler) Archive(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.Archive(c.Context(), id)
	if err != nil {
		if err.Error() == "agenda item not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"archived": true})
}
