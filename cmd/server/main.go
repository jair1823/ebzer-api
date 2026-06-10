package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"creaciones-api/internal/agenda"
	"creaciones-api/internal/db"
	"creaciones-api/internal/incomes"
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

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
	}))

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

	statusService := orders.NewStatusService(statusRepo)
	ordersService := orders.NewService(ordersRepo)

	statusHandler := orders.NewStatusHandler(statusService)
	ordersHandler := orders.NewHandler(ordersService)

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

	statusGroup := api.Group("/order-statuses")
	statusHandler.RegisterRoutes(statusGroup)

	ordersGroup := api.Group("/orders")
	ordersHandler.RegisterRoutes(ordersGroup)

	// ---------------------------------------
	// Incomes API Setup
	// ---------------------------------------
	incomesService := incomes.NewService(incomesRepo)
	incomesHandler := incomes.NewHandler(incomesService)

	incomesGroup := api.Group("/incomes")
	incomesHandler.RegisterRoutes(incomesGroup)

	// ---------------------------------------
	// Agenda API Setup
	// ---------------------------------------
	agendaRepo := agenda.NewRepository(conn)
	agendaService := agenda.NewService(agendaRepo)
	agendaHandler := agenda.NewHandler(agendaService)

	agendaGroup := api.Group("/agenda-items")
	agendaHandler.RegisterRoutes(agendaGroup)

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
