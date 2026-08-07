package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/handler"
	"book-bus/internal/middleware"
)

// Server wraps the HTTP router and its dependencies.
type Server struct {
	router     *gin.Engine
	pool       *pgxpool.Pool
	busHandler *handler.BusHandler
}

// New creates a fully wired Server: middleware applied, all routes registered.
func New(pool *pgxpool.Pool, busHandler *handler.BusHandler) *Server {
	s := &Server{
		router:     gin.New(),
		pool:       pool,
		busHandler: busHandler,
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

// Run starts the HTTP server on the given port (e.g. "8080").
func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}

// setupMiddleware registers all global middleware.
func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.RequestLogger())
}

// setupRoutes registers all application routes.
func (s *Server) setupRoutes() {
	// Health check — lives outside versioned API group
	s.router.GET("/health", s.healthCheck)

	// API v1
	v1 := s.router.Group("/api/v1")
	s.busHandler.RegisterRoutes(v1)
}

// healthCheck pings the DB and returns the service status.
func (s *Server) healthCheck(c *gin.Context) {
	if err := s.pool.Ping(c.Request.Context()); err != nil {
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
}
