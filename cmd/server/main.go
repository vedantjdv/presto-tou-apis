package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/presto-tou-apis/docs"
	"github.com/presto-tou-apis/internal/db"
	"github.com/presto-tou-apis/internal/handlers"
	"github.com/presto-tou-apis/internal/repository"
	"github.com/presto-tou-apis/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title EV Charger TOU Pricing API
// @version 1.0
// @description A backend service for managing Time-of-Use (TOU) pricing for EV chargers.
// @host localhost:8071
// @BasePath /

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Connect to database
	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize schema
	if err := db.InitSchema(pool); err != nil {
		log.Fatalf("Could not initialize schema: %v", err)
	}

	// Initialize components
	repo := repository.NewRepository(pool)
	svc := service.NewService(repo)
	h := handlers.NewHandler(svc)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes
	v1 := r.Group("/v1")
	{
		v1.POST("/chargers", h.CreateCharger)
		v1.GET("/chargers/:id/price", h.GetPrice)
		v1.POST("/chargers/assign-schedule", h.AssignSchedule)
		v1.GET("/timezones", h.ListTimezones)

		v1.POST("/schedules", h.CreateSchedule)
		v1.GET("/schedules", h.ListSchedules)
		v1.PUT("/schedules/:id", h.UpdateSchedule)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8071"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
