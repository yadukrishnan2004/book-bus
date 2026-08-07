package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
	"book-bus/internal/models"
)

// ScheduleHandler handles HTTP requests for schedule operations.
type ScheduleHandler struct {
	svc domain.ScheduleService
}

// NewScheduleHandler creates a new ScheduleHandler.
func NewScheduleHandler(svc domain.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

// RegisterRoutes wires all schedule endpoints onto the given router group.
func (h *ScheduleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	schedules := rg.Group("/schedules")
	{
		schedules.POST("", h.Create)
		schedules.GET("", h.List)
		schedules.GET("/:id", h.GetByID)
		schedules.GET("/:id/seats", h.GetSeatMap) // Step 2 of booking flow
	}
}

// Create handles POST /api/v1/schedules
func (h *ScheduleHandler) Create(c *gin.Context) {
	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	schedule, err := h.svc.CreateSchedule(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		if errors.Is(err, domain.ErrArrivalBeforeDeparture) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "schedule created successfully",
		"data":    models.NewScheduleResponse(schedule),
	})
}

// GetByID handles GET /api/v1/schedules/:id
func (h *ScheduleHandler) GetByID(c *gin.Context) {
	schedule, err := h.svc.GetSchedule(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.NewScheduleResponse(schedule)})
}

// List handles GET /api/v1/schedules?origin=&destination=&date=YYYY-MM-DD&limit=20&offset=0
func (h *ScheduleHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	filter := domain.ScheduleFilter{
		Origin:      c.Query("origin"),
		Destination: c.Query("destination"),
		Date:        c.Query("date"),
	}

	schedules, err := h.svc.ListSchedules(c.Request.Context(), filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schedules"})
		return
	}

	response := make([]*models.ScheduleResponse, 0, len(schedules))
	for _, s := range schedules {
		response = append(response, models.NewScheduleResponse(s))
	}
	c.JSON(http.StatusOK, gin.H{"data": response, "count": len(response), "limit": limit, "offset": offset})
}

// GetSeatMap handles GET /api/v1/schedules/:id/seats  (Step 2 of booking flow)
func (h *ScheduleHandler) GetSeatMap(c *gin.Context) {
	scheduleID := c.Param("id")

	seats, err := h.svc.GetSeatMap(c.Request.Context(), scheduleID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get seat map"})
		return
	}

	// Count available seats
	available := 0
	seatItems := make([]models.SeatItem, len(seats))
	for i, s := range seats {
		seatItems[i] = models.SeatItem{Number: s.Number, IsAvailable: s.IsAvailable}
		if s.IsAvailable {
			available++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": models.SeatMapResponse{
			ScheduleID:     scheduleID,
			TotalSeats:     len(seats),
			AvailableSeats: available,
			Seats:          seatItems,
		},
	})
}
