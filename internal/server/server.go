package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/domain"
	"book-bus/internal/handler"
	"book-bus/internal/middleware"
)

// Server wraps the HTTP router and its dependencies.
type Server struct {
	router          *gin.Engine
	pool            *pgxpool.Pool
	authSvc         domain.AuthService
	authHandler     *handler.AuthHandler
	busHandler      *handler.BusHandler
	routeHandler    *handler.RouteHandler
	scheduleHandler *handler.ScheduleHandler
	bookingHandler  *handler.BookingHandler
}

// New creates a fully wired Server: middleware applied, all routes registered.
func New(
	pool *pgxpool.Pool,
	authSvc domain.AuthService,
	authHandler *handler.AuthHandler,
	busHandler *handler.BusHandler,
	routeHandler *handler.RouteHandler,
	scheduleHandler *handler.ScheduleHandler,
	bookingHandler *handler.BookingHandler,
) *Server {
	s := &Server{
		router:          gin.New(),
		pool:            pool,
		authSvc:         authSvc,
		authHandler:     authHandler,
		busHandler:      busHandler,
		routeHandler:    routeHandler,
		scheduleHandler: scheduleHandler,
		bookingHandler:  bookingHandler,
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

// HTTPServer returns a standard *http.Server instance for graceful shutdown support.
func (s *Server) HTTPServer(port string) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}
}

// setupMiddleware registers all global middleware.
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.CORS())
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.RequestLogger())
}

// setupRoutes registers all application routes.
func (s *Server) setupRoutes() {
	// Health check — lives outside versioned API group
	s.router.GET("/health", s.healthCheck)

	// API v1 root group
	v1 := s.router.Group("/api/v1")

	// Protected middleware shortcuts
	jwtAuth := middleware.JWTAuth(s.authSvc)
	optionalJwtAuth := middleware.OptionalJWTAuth(s.authSvc)
	requireAdmin := middleware.RequireRole(string(domain.UserRoleAdmin))

	// Auth routes (Public registration/login + Protected profile)
	protectedV1 := v1.Group("")
	protectedV1.Use(jwtAuth)
	s.authHandler.RegisterRoutes(v1, protectedV1)

	// Public Schedule routes (Search, View Schedule, View Seat Map)
	publicSchedules := v1.Group("/schedules")
	{
		publicSchedules.GET("", s.scheduleHandler.List)
		publicSchedules.GET("/:id", s.scheduleHandler.GetByID)
		publicSchedules.GET("/:id/seats", s.scheduleHandler.GetSeatMap)
	}

	// Booking routes
	bookingsGroup := v1.Group("/bookings")
	{
		// Preview & Confirm support optional auth (binds user_id if logged in)
		bookingsGroup.POST("/preview", optionalJwtAuth, s.bookingHandler.Preview)
		bookingsGroup.POST("", optionalJwtAuth, s.bookingHandler.Confirm)
		
		// My bookings requires authentication
		bookingsGroup.GET("/my-bookings", jwtAuth, s.bookingHandler.GetMyBookings)

		// Public reference lookup & cancellation
		bookingsGroup.GET("/:reference", optionalJwtAuth, s.bookingHandler.GetByReference)
		bookingsGroup.POST("/:reference/cancel", optionalJwtAuth, s.bookingHandler.Cancel)
	}

	// Admin-protected Management APIs
	adminGroup := v1.Group("")
	adminGroup.Use(jwtAuth, requireAdmin)
	{
		// Bus Management
		adminBuses := adminGroup.Group("/buses")
		{
			adminBuses.POST("", s.busHandler.Create)
			adminBuses.GET("", s.busHandler.List)
			adminBuses.GET("/:id", s.busHandler.GetByID)
		}

		// Route Management
		adminRoutes := adminGroup.Group("/routes")
		{
			adminRoutes.POST("", s.routeHandler.Create)
			adminRoutes.GET("", s.routeHandler.List)
		}

		// Schedule Management (Create schedule & Update trip status)
		adminSchedules := adminGroup.Group("/schedules")
		{
			adminSchedules.POST("", s.scheduleHandler.Create)
			adminSchedules.PATCH("/:id/status", s.scheduleHandler.UpdateStatus)
		}
	}
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
