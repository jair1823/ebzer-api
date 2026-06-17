package expenses

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

func (h *Handler) Create(c *fiber.Ctx) error {
	var dto CreateExpenseDTO
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
			EntityType: "expenses",
			EntityID:   id,
			Action:     "create",
			Actor:      actorFromContext(c),
			Summary:    "Gasto creado",
			After:      after,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	fromPtr := queryPtr(c, "from")
	toPtr := queryPtr(c, "to")
	comercioIDPtr := queryPtr(c, "comercio_id")

	expenses, err := h.service.GetAll(c.Context(), fromPtr, toPtr, comercioIDPtr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(expenses)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	expense, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if expense == nil {
		return fiber.NewError(fiber.StatusNotFound, "expense not found")
	}

	return c.JSON(expense)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	before, _ := h.service.GetByID(c.Context(), id)

	var dto UpdateExpenseDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.Update(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "expense not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "expenses",
			EntityID:   id,
			Action:     "update",
			Actor:      actorFromContext(c),
			Summary:    "Gasto actualizado",
			Before:     before,
			After:      after,
		})
	}

	return c.JSON(fiber.Map{"updated": true})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	before, _ := h.service.GetByID(c.Context(), id)
	actor := actorFromContext(c)

	err := h.service.Delete(c.Context(), id, actor.UserID)
	if err != nil {
		if err.Error() == "expense not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if h.audit != nil {
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "expenses",
			EntityID:   id,
			Action:     "delete",
			Actor:      actor,
			Summary:    "Gasto eliminado",
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

func (h *Handler) CreateComercio(c *fiber.Ctx) error {
	var dto CreateComercioDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := h.service.CreateComercio(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetComercios(c *fiber.Ctx) error {
	comercios, err := h.service.GetComercios(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(comercios)
}

func (h *Handler) UpdateComercio(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateComercioDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.UpdateComercio(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "comercio not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

func (h *Handler) DeleteComercio(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteComercio(c.Context(), id)
	if err != nil {
		if err.Error() == "comercio not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}

func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	var dto CreateProductDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := h.service.CreateProduct(c.Context(), dto)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) GetProducts(c *fiber.Ctx) error {
	products, err := h.service.GetProducts(c.Context(), queryPtr(c, "comercio_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(products)
}

func (h *Handler) UpdateProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var dto UpdateProductDTO
	if err := c.BodyParser(&dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.service.UpdateProduct(c.Context(), id, dto)
	if err != nil {
		if err.Error() == "product not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"updated": true})
}

func (h *Handler) DeleteProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteProduct(c.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"deleted": true})
}

func queryPtr(c *fiber.Ctx, name string) *string {
	value := c.Query(name)
	if value == "" {
		return nil
	}
	return &value
}
