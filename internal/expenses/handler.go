package expenses

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
	// Categories sub-routes (must come BEFORE /:id to avoid conflicts)
	router.Post("/categories", h.CreateCategory)
	router.Get("/categories", h.GetAllCategories)
	router.Get("/categories/:id", h.GetCategoryByID)
	router.Put("/categories/:id", h.UpdateCategory)
	router.Delete("/categories/:id", h.DeleteCategory)

	// Expenses routes (order-specific route before generic /:id)
	router.Post("/", h.CreateExpense)
	router.Get("/", h.GetAllExpenses)
	router.Get("/order/:orderId", h.GetExpensesByOrderID)
	router.Get("/:id", h.GetExpenseByID)
	router.Put("/:id", h.UpdateExpense)
	router.Delete("/:id", h.DeleteExpense)
}

// -------------------- EXPENSE HANDLERS --------------------

func (h *Handler) CreateExpense(c *fiber.Ctx) error {
	var dto CreateExpenseDTO

	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := h.service.CreateExpense(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetAllExpenses(c *fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	category := c.Query("category")
	expenseType := c.Query("type")

	var fromPtr, toPtr, categoryPtr, typePtr *string
	if from != "" {
		fromPtr = &from
	}
	if to != "" {
		toPtr = &to
	}
	if category != "" {
		categoryPtr = &category
	}
	if expenseType != "" {
		typePtr = &expenseType
	}

	expenses, err := h.service.GetAllExpenses(c.Context(), fromPtr, toPtr, categoryPtr, typePtr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(expenses)
}

func (h *Handler) GetExpenseByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	expense, err := h.service.GetExpenseByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if expense == nil {
		return fiber.NewError(404, "expense not found")
	}

	return c.JSON(expense)
}

func (h *Handler) GetExpensesByOrderID(c *fiber.Ctx) error {
	orderId, _ := strconv.Atoi(c.Params("orderId"))

	expenses, err := h.service.GetExpensesByOrderID(c.Context(), orderId)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(expenses)
}

func (h *Handler) UpdateExpense(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateExpenseDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.UpdateExpense(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "expense not found" {
			return fiber.NewError(404, err.Error())
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

func (h *Handler) DeleteExpense(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteExpense(c.Context(), id)
	if err != nil {
		if err.Error() == "expense not found" {
			return fiber.NewError(404, "expense not found")
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// -------------------- CATEGORY HANDLERS --------------------

func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	var dto CreateCategoryDTO

	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := h.service.CreateCategory(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetAllCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetAllCategories(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(categories)
}

func (h *Handler) GetCategoryByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	category, err := h.service.GetCategoryByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	if category == nil {
		return fiber.NewError(404, "category not found")
	}

	return c.JSON(category)
}

func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateCategoryDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.UpdateCategory(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "category not found" {
			return fiber.NewError(404, err.Error())
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

func (h *Handler) DeleteCategory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteCategory(c.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			return fiber.NewError(404, "category not found")
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}
