package handler

import (
	"errors"
	"net/http"

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
		schedules.GET("/:id/seats", h.GetSeatMap)      // Step 2 of booking flow
		schedules.PATCH("/:id/status", h.UpdateStatus) // Trip completion / lifecycle update
	}
}

// Create handles POST /api/v1/schedules
func (h *ScheduleHandler) Create(c *gin.Context) {
	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "invalid request", err.Error())
		return
	}

	schedule, err := h.svc.CreateSchedule(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		HandleError(c, err, "failed to create schedule")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "schedule created successfully",
		"data":    models.NewScheduleResponse(schedule),
	})
}

// GetByID handles GET /api/v1/schedules/:id
func (h *ScheduleHandler) GetByID(c *gin.Context) {
	schedule, err := h.svc.GetSchedule(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "schedule not found")
			return
		}
		HandleError(c, err, "failed to fetch schedule")
		return
	}
	RespondJSON(c, http.StatusOK, gin.H{"data": models.NewScheduleResponse(schedule)})
}

// List handles GET /api/v1/schedules?origin=&destination=&date=YYYY-MM-DD&limit=20&offset=0
func (h *ScheduleHandler) List(c *gin.Context) {
	limit, offset := ParsePagination(c)

	filter := domain.ScheduleFilter{
		Origin:      c.Query("origin"),
		Destination: c.Query("destination"),
		Date:        c.Query("date"),
	}

	schedules, err := h.svc.ListSchedules(c.Request.Context(), filter, limit, offset)
	if err != nil {
		HandleError(c, err, "failed to list schedules")
		return
	}

	response := make([]*models.ScheduleResponse, 0, len(schedules))
	for _, s := range schedules {
		response = append(response, models.NewScheduleResponse(s))
	}
	RespondJSON(c, http.StatusOK, gin.H{"data": response, "count": len(response), "limit": limit, "offset": offset})
}

// GetSeatMap handles GET /api/v1/schedules/:id/seats  (Step 2 of booking flow)
func (h *ScheduleHandler) GetSeatMap(c *gin.Context) {
	scheduleID := c.Param("id")

	seats, err := h.svc.GetSeatMap(c.Request.Context(), scheduleID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "schedule not found")
			return
		}
		HandleError(c, err, "failed to get seat map")
		return
	}

	available := 0
	seatItems := make([]models.SeatItem, len(seats))
	for i, s := range seats {
		seatItems[i] = models.SeatItem{Number: s.Number, IsAvailable: s.IsAvailable}
		if s.IsAvailable {
			available++
		}
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"data": models.SeatMapResponse{
			ScheduleID:     scheduleID,
			TotalSeats:     len(seats),
			AvailableSeats: available,
			Seats:          seatItems,
		},
	})
}

// UpdateStatus handles PATCH /api/v1/schedules/:id/status
func (h *ScheduleHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateScheduleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "invalid request", err.Error())
		return
	}

	schedule, affected, err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		HandleError(c, err, "failed to update schedule status")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"message": "schedule status updated successfully",
		"data": models.UpdateScheduleStatusResponse{
			ID:               schedule.ID,
			Status:           schedule.Status,
			AffectedBookings: affected,
			UpdatedAt:        schedule.UpdatedAt,
		},
	})
}
