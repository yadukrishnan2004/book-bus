package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
	"book-bus/internal/models"
)

// BusHandler handles HTTP requests for bus operations.
type BusHandler struct {
	svc domain.BusService
}

// NewBusHandler creates a new BusHandler.
func NewBusHandler(svc domain.BusService) *BusHandler {
	return &BusHandler{svc: svc}
}

// RegisterRoutes wires all bus routes onto the given router group.
func (h *BusHandler) RegisterRoutes(rg *gin.RouterGroup) {
	buses := rg.Group("/buses")
	{
		buses.POST("", h.Create)
		buses.GET("", h.List)
		buses.GET("/:id", h.GetByID)
	}
}

// Create handles POST /api/v1/buses
func (h *BusHandler) Create(c *gin.Context) {
	var req models.CreateBusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	bus, err := h.svc.CreateBus(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateKey) {
			RespondError(c, http.StatusConflict, "a bus with this number plate already exists")
			return
		}
		HandleError(c, err, "failed to register bus")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "bus registered successfully",
		"data":    models.NewBusResponse(bus),
	})
}

// GetByID handles GET /api/v1/buses/:id
func (h *BusHandler) GetByID(c *gin.Context) {
	bus, err := h.svc.GetBus(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "bus not found")
			return
		}
		HandleError(c, err, "failed to fetch bus")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"data": models.NewBusResponse(bus),
	})
}

// List handles GET /api/v1/buses?limit=20&offset=0
func (h *BusHandler) List(c *gin.Context) {
	limit, offset := ParsePagination(c)

	buses, err := h.svc.ListBuses(c.Request.Context(), limit, offset)
	if err != nil {
		HandleError(c, err, "failed to list buses")
		return
	}

	response := make([]*models.BusResponse, 0, len(buses))
	for _, b := range buses {
		response = append(response, models.NewBusResponse(b))
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"data":   response,
		"limit":  limit,
		"offset": offset,
		"count":  len(response),
	})
}
