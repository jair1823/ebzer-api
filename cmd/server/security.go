package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	defaultDevOrigins   = "http://localhost:8080,http://localhost:5173"
	defaultGeneralLimit = 120
	defaultAuthLimit    = 10
)

func corsConfigFromEnv() (cors.Config, error) {
	origins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origins == "" {
		if isProduction() {
			return cors.Config{}, errors.New("CORS_ALLOWED_ORIGINS is required in production")
		}
		origins = defaultDevOrigins
	}

	return cors.Config{
		AllowOrigins: origins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
	}, nil
}

func securityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	}
}

func rateLimiterFromEnv(envName string, fallback int) fiber.Handler {
	maxRequests := fallback
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			maxRequests = parsed
		}
	}

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests",
			})
		},
	})
}

func isProduction() bool {
	return strings.EqualFold(os.Getenv("APP_ENV"), "production") ||
		strings.EqualFold(os.Getenv("RAILWAY_ENVIRONMENT"), "production")
}
