package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"creaciones-api/internal/db"
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
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
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
	ordersService := orders.NewService(ordersRepo)
	ordersHandler := orders.NewHandler(ordersService)

	api := app.Group("/api")

	ordersGroup := api.Group("/orders")
	ordersHandler.RegisterRoutes(ordersGroup)

	// ---------------------------------------
	// Start Server
	// ---------------------------------------

	log.Println("🚀 Server running on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
