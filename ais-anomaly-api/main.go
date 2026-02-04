package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/db"
	"github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/handlers"
)

func main() {
	// Initialize database connection
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "AIS Anomaly API v1.0.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Initialize handlers
	anomalyHandler := handlers.NewAnomalyHandler(database)

	// Routes
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "ais-anomaly-api",
		})
	})

	// Anomaly group routes
	api.Get("/anomaly-groups", anomalyHandler.GetAnomalyGroups)
	api.Get("/anomaly-groups/:id", anomalyHandler.GetAnomalyGroupByID)

	// Get port from environment variable or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting AIS Anomaly API on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
