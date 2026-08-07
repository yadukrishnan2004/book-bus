package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"book-bus/internal/models"
	"book-bus/internal/repository"
)

// BusHandler holds dependencies for bus-related HTTP handlers.
type BusHandler struct {
	repo *repository.BusRepository
}

// NewBusHandler creates a new BusHandler.
func NewBusHandler(repo *repository.BusRepository) *BusHandler {
	return &BusHandler{repo: repo}
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

// Create godoc
// POST /api/v1/buses
// Registers a new bus in the system.
func (h *BusHandler) Create(c *gin.Context) {
	var req models.CreateBusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	bus, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		// Unique constraint on number_plate
		if isDuplicateKey(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "a bus with this number plate already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to register bus",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "bus registered successfully",
		"data":    toBusResponse(bus),
	})
}

// GetByID godoc
// GET /api/v1/buses/:id
// Returns a single bus by its UUID.
func (h *BusHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	bus, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "bus not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch bus",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toBusResponse(bus),
	})
}

// List godoc
// GET /api/v1/buses?limit=20&offset=0
// Returns a paginated list of all buses.
func (h *BusHandler) List(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	buses, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list buses",
		})
		return
	}

	// Return empty array instead of null when no buses exist
	response := make([]*models.BusResponse, 0, len(buses))
	for _, b := range buses {
		response = append(response, toBusResponse(b))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   response,
		"limit":  limit,
		"offset": offset,
		"count":  len(response),
	})
}

// toBusResponse converts a Bus model to the API response shape.
func toBusResponse(b *models.Bus) *models.BusResponse {
	return &models.BusResponse{
		ID:          b.ID,
		Name:        b.Name,
		NumberPlate: b.NumberPlate,
		TotalSeats:  b.TotalSeats,
		BusType:     b.BusType,
		Description: b.Description,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
	}
}

// isDuplicateKey checks if the error is a Postgres unique constraint violation (code 23505).
func isDuplicateKey(err error) bool {
	return err != nil && (containsCode(err, "23505"))
}

func containsCode(err error, code string) bool {
	return err != nil && len(err.Error()) > 0 &&
		(func() bool {
			type pgErr interface{ SQLState() string }
			var pg pgErr
			if errors.As(err, &pg) {
				return pg.SQLState() == code
			}
			return false
		})()
}
