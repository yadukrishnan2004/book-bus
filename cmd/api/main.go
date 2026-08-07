package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"book-bus/internal/config"
	"book-bus/internal/db"
	"book-bus/internal/handler"
	"book-bus/internal/repository"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Connect to the database
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("✅ Connected to PostgreSQL successfully")

	// --- Repositories ---
	busRepo := repository.NewBusRepository(pool)

	// --- Handlers ---
	busHandler := handler.NewBusHandler(busRepo)

	// --- Router ---
	g := gin.Default()

	// Health check
	g.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "database unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Bus Ticket Booking API is running!",
		})
	})

	// API v1 routes
	v1 := g.Group("/api/v1")
	busHandler.RegisterRoutes(v1)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := g.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}