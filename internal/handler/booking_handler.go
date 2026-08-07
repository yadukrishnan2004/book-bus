package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
	"book-bus/internal/models"
)

// BookingHandler handles HTTP requests for booking operations.
type BookingHandler struct {
	svc domain.BookingService
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(svc domain.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// RegisterRoutes wires all booking endpoints onto the given router group.
func (h *BookingHandler) RegisterRoutes(rg *gin.RouterGroup) {
	bookings := rg.Group("/bookings")
	{
		bookings.POST("/preview", h.Preview) // Step 3
		bookings.POST("", h.Confirm)         // Step 4
		bookings.GET("/:reference", h.GetByReference)
		bookings.POST("/:reference/cancel", h.Cancel)
	}
}

// Preview handles POST /api/v1/bookings/preview (Step 3: Validate & Price Quote)
func (h *BookingHandler) Preview(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "invalid request", err.Error())
		return
	}

	preview, err := h.svc.PreviewBooking(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		HandleError(c, err, "failed to preview booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"data": models.NewBookingPreviewResponse(preview)})
}

// Confirm handles POST /api/v1/bookings (Step 4: Lock seats & Confirm)
func (h *BookingHandler) Confirm(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "invalid request", err.Error())
		return
	}

	bookings, reference, err := h.svc.ConfirmBooking(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		HandleError(c, err, "failed to confirm booking")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "booking confirmed successfully",
		"data":    models.NewBookingConfirmResponse(bookings, reference),
	})
}

// GetByReference handles GET /api/v1/bookings/:reference
func (h *BookingHandler) GetByReference(c *gin.Context) {
	ref := c.Param("reference")

	bookings, err := h.svc.GetBooking(c.Request.Context(), ref)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "booking not found")
			return
		}
		HandleError(c, err, "failed to fetch booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"data": models.NewBookingConfirmResponse(bookings, ref)})
}

// Cancel handles POST /api/v1/bookings/:reference/cancel
func (h *BookingHandler) Cancel(c *gin.Context) {
	ref := c.Param("reference")

	err := h.svc.CancelBooking(c.Request.Context(), ref)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "booking not found")
			return
		}
		HandleError(c, err, "failed to cancel booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}
