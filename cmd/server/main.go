package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"creaciones-api/internal/agenda"
	"creaciones-api/internal/audit"
	"creaciones-api/internal/auth"
	"creaciones-api/internal/db"
	"creaciones-api/internal/expenses"
	"creaciones-api/internal/incomes"
	"creaciones-api/internal/insights"
	"creaciones-api/internal/orders"
)

func main() {
	// ---------------------------------------
	// DB Connection
	// ---------------------------------------

	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close()

	// ---------------------------------------
	// Run Migrations
	// ---------------------------------------

	log.Println("🔄 Running database migrations...")
	if err := db.RunMigrations(conn, "internal/db/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Migrations completed successfully")

	// ---------------------------------------
	// Auth Setup
	// ---------------------------------------

	authConfig, err := auth.LoadConfig()
	if err != nil {
		log.Fatalf("Invalid auth configuration: %v", err)
	}
	authRepo := auth.NewRepository(conn)
	jwtService := auth.NewJWTService(authConfig.JWTSecret)
	authService := auth.NewService(authRepo, jwtService, authConfig)
	authHandler := auth.NewHandler(authService)

	if err := authService.BootstrapInitialAdmin(context.Background()); err != nil {
		log.Fatalf("Failed to bootstrap initial admin: %v", err)
	}

	// ---------------------------------------
	// Fiber Config
	// ---------------------------------------

	app := fiber.New(fiber.Config{
		AppName:      "Creaciones API",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	})

	// ---------------------------------------
	// Middlewares
	// ---------------------------------------

	app.Use(logger.New()) // Logs every request
	app.Use(securityHeaders())

	corsConfig, err := corsConfigFromEnv()
	if err != nil {
		log.Fatalf("Invalid CORS configuration: %v", err)
	}
	app.Use(cors.New(corsConfig))

	// ---------------------------------------
	// Health Check
	// ---------------------------------------

	// Simple ping route
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	})

	//Database ping endpoint
	app.Get("/dbping", func(c *fiber.Ctx) error {
		// Intentar conectar con reintentos (máx 3 intentos, 2 segundos entre intentos)
		connection, err := db.ConnectWithRetry(3, 2*time.Second)
		if err != nil {
			log.Printf("Error de conexión: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error":   "Database connection failed",
				"details": err.Error(),
			})
		}

		defer connection.Close()

		return c.JSON(fiber.Map{
			"message": "Database connection successful",
		})
	})

	// ---------------------------------------
	// Orders API Setup
	// ---------------------------------------

	ordersRepo := orders.NewRepository(conn)
	statusRepo := orders.NewStatusRepository(conn)
	incomesRepo := incomes.NewRepository(conn)
	expensesRepo := expenses.NewRepository(conn)
	auditRepo := audit.NewRepository(conn)
	insightsRepo := insights.NewRepository(conn)

	statusService := orders.NewStatusService(statusRepo)
	ordersService := orders.NewService(ordersRepo)

	statusHandler := orders.NewStatusHandler(statusService)
	ordersHandler := orders.NewHandler(ordersService, auditRepo)
	auditHandler := audit.NewHandler(auditRepo)
	insightsHandler := insights.NewHandler(insights.NewService(insightsRepo))

	api := app.Group("/api")
	// API key guard — set API_KEY env var to enable; skipped if unset (dev mode)
	apiKey := os.Getenv("API_KEY")
	if apiKey != "" {
		api.Use(func(c *fiber.Ctx) error {
			key := c.Get("X-API-Key")
			if key != apiKey {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized",
				})
			}
			return c.Next()
		})
	}
	api.Use(rateLimiterFromEnv("RATE_LIMIT_GENERAL", defaultGeneralLimit))

	authGroup := api.Group("/auth")
	authGroup.Use(rateLimiterFromEnv("RATE_LIMIT_AUTH", defaultAuthLimit))
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)

	api.Use(auth.RequireAuth(jwtService, authRepo))

	authGroup.Get("/me", authHandler.Me)
	authGroup.Post("/logout", authHandler.Logout)

	usersGroup := api.Group("/users", auth.RequireRole(auth.RoleAdmin))
	usersGroup.Get("/", authHandler.ListUsers)
	usersGroup.Post("/", authHandler.CreateUser)
	usersGroup.Get("/:id", authHandler.GetUser)
	usersGroup.Put("/:id", authHandler.UpdateUser)
	usersGroup.Delete("/:id", authHandler.DeactivateUser)

	auditGroup := api.Group("/audit-events", auth.RequireRole(auth.RoleAdmin))
	auditGroup.Get("/", auditHandler.GetAll)

	insightsGroup := api.Group("/insights")
	insightsGroup.Get("/summary", insightsHandler.Summary)

	statusGroup := api.Group("/order-statuses")
	statusGroup.Put("/reorder", auth.RequireRole(auth.RoleAdmin), statusHandler.Reorder)
	statusGroup.Get("/", statusHandler.GetAll)
	statusGroup.Post("/", auth.RequireRole(auth.RoleAdmin), statusHandler.Create)
	statusGroup.Get("/:id", statusHandler.GetByID)
	statusGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin), statusHandler.Update)
	statusGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin), statusHandler.Deactivate)

	ordersGroup := api.Group("/orders")
	ordersGroup.Post("/", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), ordersHandler.Create)
	ordersGroup.Get("/", ordersHandler.GetAll)
	ordersGroup.Get("/:id", ordersHandler.GetByID)
	ordersGroup.Get("/:id/payment-status", ordersHandler.GetPaymentStatus)
	ordersGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), ordersHandler.Update)
	ordersGroup.Post("/:id/finish", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), ordersHandler.FinishOrder)
	ordersGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), ordersHandler.Delete)

	// ---------------------------------------
	// Incomes API Setup
	// ---------------------------------------
	incomesService := incomes.NewService(incomesRepo)
	incomesHandler := incomes.NewHandler(incomesService, auditRepo)

	incomesGroup := api.Group("/incomes")
	incomesGroup.Post("/", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), incomesHandler.Create)
	incomesGroup.Get("/", incomesHandler.GetAll)
	incomesGroup.Get("/:id", incomesHandler.GetByID)
	incomesGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin), incomesHandler.Update)
	incomesGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin), incomesHandler.Delete)

	// ---------------------------------------
	// Expenses API Setup
	// ---------------------------------------
	expensesService := expenses.NewService(expensesRepo)
	expensesHandler := expenses.NewHandler(expensesService, auditRepo)

	expensesGroup := api.Group("/expenses")
	expensesGroup.Post("/", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), expensesHandler.Create)
	expensesGroup.Get("/", expensesHandler.GetAll)
	expensesGroup.Get("/:id", expensesHandler.GetByID)
	expensesGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.Update)
	expensesGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.Delete)

	comerciosGroup := api.Group("/comercios")
	comerciosGroup.Get("/", expensesHandler.GetComercios)
	comerciosGroup.Post("/", auth.RequireRole(auth.RoleAdmin), expensesHandler.CreateComercio)
	comerciosGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.UpdateComercio)
	comerciosGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.DeleteComercio)

	productsGroup := api.Group("/products")
	productsGroup.Get("/", expensesHandler.GetProducts)
	productsGroup.Post("/", auth.RequireRole(auth.RoleAdmin), expensesHandler.CreateProduct)
	productsGroup.Put("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.UpdateProduct)
	productsGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin), expensesHandler.DeleteProduct)

	// ---------------------------------------
	// Agenda API Setup
	// ---------------------------------------
	agendaRepo := agenda.NewRepository(conn)
	agendaService := agenda.NewService(agendaRepo)
	agendaHandler := agenda.NewHandler(agendaService, auditRepo)

	agendaGroup := api.Group("/agenda-items")
	agendaGroup.Post("/", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), agendaHandler.Create)
	agendaGroup.Get("/", agendaHandler.GetAll)
	agendaGroup.Get("/:id", agendaHandler.GetByID)
	agendaGroup.Patch("/:id", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), agendaHandler.Update)
	agendaGroup.Delete("/:id", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), agendaHandler.Delete)
	agendaGroup.Patch("/:id/complete", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), agendaHandler.Complete)
	agendaGroup.Patch("/:id/archive", auth.RequireRole(auth.RoleAdmin, auth.RoleOperator), agendaHandler.Archive)

	// ---------------------------------------
	// Start Server
	// ---------------------------------------

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server running on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
