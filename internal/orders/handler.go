package orders

import (
	"creaciones-api/internal/audit"
	"creaciones-api/internal/auth"
	"strconv"
	"strings"

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
	router.Get("/:id/payment-status", h.GetPaymentStatus)
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
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "orders",
			EntityID:   id,
			Action:     "create",
			Actor:      actorFromContext(c),
			Summary:    "Pedido creado",
			After:      after,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// -------------------- GET ALL --------------------

func (h *Handler) GetAll(c *fiber.Ctx) error {
	filter, err := parseOrderFilter(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	orders, err := h.service.GetAll(c.Context(), filter)
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
	before, _ := h.service.GetByID(c.Context(), id)

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
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "orders",
			EntityID:   id,
			Action:     "update",
			Actor:      actorFromContext(c),
			Summary:    "Pedido actualizado",
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
		if err.Error() == "order not found" {
			return fiber.NewError(404, "order not found")
		}
		return fiber.NewError(500, err.Error())
	}
	if h.audit != nil {
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "orders",
			EntityID:   id,
			Action:     "delete",
			Actor:      actor,
			Summary:    "Pedido eliminado",
			Before:     before,
		})
	}

	return c.JSON(fiber.Map{"deleted": true})
}

// -------------------- FINISH ORDER --------------------
func (h *Handler) FinishOrder(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	before, _ := h.service.GetByID(c.Context(), id)
	result, err := h.service.FinishOrder(c.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			return fiber.NewError(404, "order not found")
		}
		return fiber.NewError(500, err.Error())
	}
	if h.audit != nil {
		after, _ := h.service.GetByID(c.Context(), id)
		_ = h.audit.Record(c.Context(), audit.CreateEventDTO{
			EntityType: "orders",
			EntityID:   id,
			Action:     "finish",
			Actor:      actorFromContext(c),
			Summary:    "Pedido finalizado y pago registrado",
			Before:     before,
			After:      after,
		})
	}
	return c.JSON(result)
}

// -------------------- GET PAYMENT STATUS --------------------
func (h *Handler) GetPaymentStatus(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	paymentStatus, err := h.service.GetPaymentStatus(c.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			return fiber.NewError(404, "order not found")
		}
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(paymentStatus)
}

func parseOrderFilter(c *fiber.Ctx) (OrderFilterDTO, error) {
	filter := OrderFilterDTO{
		From:         queryPtr(c, "from"),
		To:           queryPtr(c, "to"),
		Search:       queryPtr(c, "search"),
		DeliveryFrom: queryPtr(c, "delivery_from"),
		DeliveryTo:   queryPtr(c, "delivery_to"),
	}

	if statusIDStr := c.Query("status_id"); statusIDStr != "" {
		id, err := strconv.Atoi(statusIDStr)
		if err != nil {
			return filter, err
		}
		filter.StatusID = &id
	}

	if statusIDsStr := c.Query("status_ids"); statusIDsStr != "" {
		for _, raw := range strings.Split(statusIDsStr, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := strconv.Atoi(raw)
			if err != nil {
				return filter, err
			}
			filter.StatusIDs = append(filter.StatusIDs, id)
		}
	}

	if platformStr := c.Query("platform"); platformStr != "" {
		platform := Platform(platformStr)
		if platform != PlatformWhatsApp && platform != PlatformInstagram && platform != PlatformFacebook {
			return filter, fiber.NewError(fiber.StatusBadRequest, "invalid platform")
		}
		filter.Platform = &platform
	}

	if paymentStr := c.Query("payment_status"); paymentStr != "" {
		paymentStatus := PaymentStatusFilter(paymentStr)
		if paymentStatus != PaymentStatusFilterUnpaid && paymentStatus != PaymentStatusFilterPartial && paymentStatus != PaymentStatusFilterPaid {
			return filter, fiber.NewError(fiber.StatusBadRequest, "invalid payment_status")
		}
		filter.PaymentStatus = &paymentStatus
	}

	if overdueStr := c.Query("overdue"); overdueStr != "" {
		overdue, err := strconv.ParseBool(overdueStr)
		if err != nil {
			return filter, err
		}
		filter.Overdue = &overdue
	}

	if amountMinStr := c.Query("amount_min"); amountMinStr != "" {
		amount, err := strconv.ParseFloat(amountMinStr, 64)
		if err != nil {
			return filter, err
		}
		filter.AmountMin = &amount
	}

	if amountMaxStr := c.Query("amount_max"); amountMaxStr != "" {
		amount, err := strconv.ParseFloat(amountMaxStr, 64)
		if err != nil {
			return filter, err
		}
		filter.AmountMax = &amount
	}

	return filter, nil
}

func queryPtr(c *fiber.Ctx, name string) *string {
	value := c.Query(name)
	if value == "" {
		return nil
	}
	return &value
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
