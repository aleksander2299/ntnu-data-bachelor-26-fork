package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	"github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/db"
	_ "github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/docs"
	"github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/handlers"
)

// @title AIS Anomaly Detection API
// @version 1.0
// @description API for querying AIS anomaly detection data
// @host localhost:3000
// @BasePath /api/v1
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

	// Swagger endpoint - optionally protected with basic auth
	swaggerAuthEnabled := os.Getenv("SWAGGER_AUTH_ENABLED") == "true"

	if swaggerAuthEnabled {
		swaggerUser := os.Getenv("SWAGGER_USER")
		if swaggerUser == "" {
			swaggerUser = "admin"
		}
		swaggerPassword := os.Getenv("SWAGGER_PASSWORD")
		if swaggerPassword == "" {
			swaggerPassword = "admin"
		}

		log.Printf("Swagger basic auth enabled - User: %s", swaggerUser)
		swaggerGroup := app.Group("/swagger", basicauth.New(basicauth.Config{
			Users: map[string]string{
				swaggerUser: swaggerPassword,
			},
		}))
		swaggerGroup.Get("/*", swagger.HandlerDefault)
	} else {
		log.Println("Swagger basic auth disabled")
		app.Get("/swagger", func(c *fiber.Ctx) error {
			return c.Redirect("/swagger/index.html")
		})
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Initialize handlers
	anomalyHandler := handlers.NewAnomalyHandler(database)
	baseStationHandler := handlers.NewBaseStationHandler()

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

	api.Get("/anomaly-groups/mmsi/:mmsi", anomalyHandler.GetAnomalyGroupsByMMSI)
	api.Get("/anomaly-groups/type/:type", anomalyHandler.GetAnomalyGroupsByType)

	// Id needs to be last
	api.Get("/anomaly-groups/:id", anomalyHandler.GetAnomalyGroupByID)

	// Base station routes
	api.Get("/base-stations", baseStationHandler.GetBaseStations)
	api.Get("/base-stations/:id", baseStationHandler.GetBaseStationByID)

	// Get port from environment variable or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting AIS Anomaly API on port %s", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)
	log.Fatal(app.Listen(":" + port))
}
