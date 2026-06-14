package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"creaciones-api/internal/db"
	"creaciones-api/internal/expenses"
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
		AllowOrigins:     "http://localhost:5173, http://127.0.0.1:5173, http://192.168.1.45:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
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
	incomesRepo := incomes.NewRepository(conn)
	ordersService := orders.NewService(ordersRepo, incomesRepo)
	ordersHandler := orders.NewHandler(ordersService)

	api := app.Group("/api")

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
	// Expenses API Setup
	// ---------------------------------------
	expensesRepo := expenses.NewRepository(conn)
	expensesService := expenses.NewService(expensesRepo)
	expensesHandler := expenses.NewHandler(expensesService)

	expensesGroup := api.Group("/expenses")
	expensesHandler.RegisterRoutes(expensesGroup)

	// ---------------------------------------
	// Start Server
	// ---------------------------------------

	log.Println("🚀 Server starting...")
	log.Println("   Local:   http://localhost:3000")
	log.Println("   Network: http://0.0.0.0:3000")
	log.Println("   Access from other devices using your machine's IP address")
	
	if err := app.Listen("0.0.0.0:3000"); err != nil {
		log.Fatal(err)
	}
}
