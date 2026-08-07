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

// BusHandler handles HTTP requests for bus operations.
// It depends on the domain.BusService interface — never on a concrete type.
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	bus, err := h.svc.CreateBus(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrDuplicateKey):
			c.JSON(http.StatusConflict, gin.H{
				"error": "a bus with this number plate already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to register bus",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "bus registered successfully",
		"data":    models.NewBusResponse(bus),
	})
}

// GetByID handles GET /api/v1/buses/:id
func (h *BusHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	bus, err := h.svc.GetBus(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "bus not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch bus",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": models.NewBusResponse(bus),
	})
}

// List handles GET /api/v1/buses?limit=20&offset=0
func (h *BusHandler) List(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	buses, err := h.svc.ListBuses(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list buses",
		})
		return
	}

	response := make([]*models.BusResponse, 0, len(buses))
	for _, b := range buses {
		response = append(response, models.NewBusResponse(b))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   response,
		"limit":  limit,
		"offset": offset,
		"count":  len(response),
	})
}
