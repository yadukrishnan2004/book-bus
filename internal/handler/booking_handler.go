package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
	"book-bus/internal/middleware"
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
		bookings.POST("/preview", h.Preview)
		bookings.POST("", h.Confirm)
		bookings.GET("/my-bookings", h.GetMyBookings)
		bookings.GET("/:reference", h.GetByReference)
		bookings.POST("/:reference/cancel", h.Cancel)
	}
}

// Preview handles POST /api/v1/bookings/preview
func (h *BookingHandler) Preview(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	input := req.ToDomainInput()
	if userID, ok := middleware.GetUserID(c); ok {
		input.UserID = &userID
	}

	preview, err := h.svc.PreviewBooking(c.Request.Context(), input)
	if err != nil {
		HandleError(c, err, "failed to preview booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"data": models.NewBookingPreviewResponse(preview)})
}

// Confirm handles POST /api/v1/bookings
func (h *BookingHandler) Confirm(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	input := req.ToDomainInput()
	if userID, ok := middleware.GetUserID(c); ok {
		input.UserID = &userID
	}

	bookings, reference, err := h.svc.ConfirmBooking(c.Request.Context(), input)
	if err != nil {
		HandleError(c, err, "failed to confirm booking")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "booking confirmed successfully",
		"data":    models.NewBookingConfirmResponse(bookings, reference),
	})
}

// GetMyBookings handles GET /api/v1/bookings/my-bookings
func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		RespondError(c, http.StatusUnauthorized, "authentication required to view personal bookings")
		return
	}

	bookings, err := h.svc.GetUserBookings(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, "failed to fetch user bookings")
		return
	}

	// Group bookings by reference
	grouped := make(map[string][]*domain.Booking)
	var order []string
	for _, b := range bookings {
		if _, exists := grouped[b.BookingReference]; !exists {
			order = append(order, b.BookingReference)
		}
		grouped[b.BookingReference] = append(grouped[b.BookingReference], b)
	}

	response := make([]*models.BookingConfirmResponse, 0, len(order))
	for _, ref := range order {
		response = append(response, models.NewBookingConfirmResponse(grouped[ref], ref))
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"data":  response,
		"count": len(response),
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

	// Bug 6 fix: if the booking belongs to a registered user, ensure only that
	// user or an admin can view it. Guest bookings (no user_id) are public.
	if err := enforceBookingOwnership(c, bookings[0].UserID); err != nil {
		RespondError(c, http.StatusForbidden, "you do not have permission to view this booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"data": models.NewBookingConfirmResponse(bookings, ref)})
}

// Cancel handles POST /api/v1/bookings/:reference/cancel
func (h *BookingHandler) Cancel(c *gin.Context) {
	ref := c.Param("reference")

	// Fetch the booking first so we can check ownership before mutating state.
	bookings, err := h.svc.GetBooking(c.Request.Context(), ref)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "booking not found")
			return
		}
		HandleError(c, err, "failed to fetch booking for cancellation")
		return
	}

	// Bug 6 fix: enforce ownership before allowing cancellation.
	if err := enforceBookingOwnership(c, bookings[0].UserID); err != nil {
		RespondError(c, http.StatusForbidden, "you do not have permission to cancel this booking")
		return
	}

	if err := h.svc.CancelBooking(c.Request.Context(), ref); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "booking not found")
			return
		}
		HandleError(c, err, "failed to cancel booking")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

// enforceBookingOwnership returns an error if the requester is not the booking
// owner and is not an admin. Returns nil for guest bookings (bookingUserID == nil).
func enforceBookingOwnership(c *gin.Context, bookingUserID *string) error {
	if bookingUserID == nil {
		// Guest booking — no ownership to enforce.
		return nil
	}
	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		// No auth token provided for a user-owned booking.
		return errors.New("unauthorized")
	}
	// Admins can access any booking.
	if roleVal, exists := c.Get(middleware.CtxRoleKey); exists {
		if role, ok := roleVal.(string); ok && role == "admin" {
			return nil
		}
	}
	if requesterID != *bookingUserID {
		return errors.New("forbidden")
	}
	return nil
}

