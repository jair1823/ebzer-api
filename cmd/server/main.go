package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"creaciones-api/internal/db"
)

func main() {
	app := fiber.New()

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

	log.Println("🚀 Server running on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
