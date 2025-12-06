package incomes

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
	router.Delete("/:id", h.Delete)
}

// -------------------- CREATE --------------------

func (h *Handler) Create(c *fiber.Ctx) error {
	var dto CreateIncomeDTO

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
	from := c.Query("from")
	to := c.Query("to")

	// only pass pointers if the values are not empty
	var fromPtr *string
	if from != "" {
		fromPtr = &from
	}

	var toPtr *string
	if to != "" {
		toPtr = &to
	}

	incomes, err := h.service.GetAll(c.Context(), fromPtr, toPtr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(incomes)
}

// -------------------- GET BY ID --------------------

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	income, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if income == nil {
		return fiber.NewError(404, "income not found")
	}

	return c.JSON(income)
}

// -------------------- UPDATE --------------------

func (h *Handler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateIncomeDTO
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
		if err.Error() == "income not found" {
			return fiber.NewError(404, "income not found")
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}
