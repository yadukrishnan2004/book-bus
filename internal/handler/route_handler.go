package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"book-bus/internal/domain"
	"book-bus/internal/models"
)

// RouteHandler handles HTTP requests for route operations.
type RouteHandler struct {
	svc domain.RouteService
}

// NewRouteHandler creates a new RouteHandler.
func NewRouteHandler(svc domain.RouteService) *RouteHandler {
	return &RouteHandler{svc: svc}
}

// RegisterRoutes wires all route endpoints onto the given router group.
func (h *RouteHandler) RegisterRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/routes")
	{
		routes.POST("", h.Create)
		routes.GET("", h.List)
	}
}

// Create handles POST /api/v1/routes
func (h *RouteHandler) Create(c *gin.Context) {
	var req models.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	route, err := h.svc.CreateRoute(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		HandleError(c, err, "failed to create route")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "route created successfully",
		"data":    models.NewRouteResponse(route),
	})
}

// List handles GET /api/v1/routes
func (h *RouteHandler) List(c *gin.Context) {
	routes, err := h.svc.ListRoutes(c.Request.Context())
	if err != nil {
		HandleError(c, err, "failed to list routes")
		return
	}

	response := make([]*models.RouteResponse, 0, len(routes))
	for _, r := range routes {
		response = append(response, models.NewRouteResponse(r))
	}

	RespondJSON(c, http.StatusOK, gin.H{"data": response, "count": len(response)})
}
