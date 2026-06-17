package incomes

import (
	"creaciones-api/internal/audit"
	"creaciones-api/internal/auth"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
	audit   audit.Repository
}

func NewHandler(s Service, auditRepo ...audit.Repository) *Handler {
	handler := &Handler{service: s}
	if len(auditRepo) > 0 {
		handler.audit = auditRepo[0]
	}
	return handler
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
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "income",
			EntityID:   id,
			Action:     "create",
			Actor:      actorFromContext(c),
			Summary:    "Ingreso creado",
			After:      after,
		})
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
	before, _ := h.service.GetByID(c.Context(), id)

	var dto UpdateIncomeDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.Update(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "income not found" {
			return fiber.NewError(404, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "income",
			EntityID:   id,
			Action:     "update",
			Actor:      actorFromContext(c),
			Summary:    "Ingreso actualizado",
			Before:     before,
			After:      after,
		})
	}

	return c.JSON(fiber.Map{"updated": true})
}

// -------------------- DELETE --------------------

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	before, _ := h.service.GetByID(c.Context(), id)
	actor := actorFromContext(c)

	err := h.service.Delete(c.Context(), id, actor.UserID)
	if err != nil {
		if err.Error() == "income not found" {
			return fiber.NewError(404, "income not found")
		}
		return fiber.NewError(500, err.Error())
	}
	if h.audit != nil {
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "income",
			EntityID:   id,
			Action:     "delete",
			Actor:      actor,
			Summary:    "Ingreso eliminado",
			Before:     before,
		})
	}

	return c.JSON(fiber.Map{"deleted": true})
}

func actorFromContext(c *fiber.Ctx) audit.Actor {
	user, ok := auth.CurrentUser(c)
	if !ok || user == nil {
		return audit.Actor{}
	}
	id := user.ID
	username := user.Username
	return audit.Actor{UserID: &id, Username: &username}
}
