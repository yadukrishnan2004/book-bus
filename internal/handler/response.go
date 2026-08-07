package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
)

// ParsePagination extracts and validates limit & offset query parameters from the request.
func ParsePagination(c *gin.Context) (limit, offset int) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

// RespondJSON sends a JSON response with a given status code.
func RespondJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// RespondError sends a standardized JSON error response.
func RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// RespondBadRequest sends a 400 Bad Request error response with details.
func RespondBadRequest(c *gin.Context, message string, details string) {
	if details != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": message, "details": details})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// HandleError maps common domain & app errors to standardized HTTP responses.
func HandleError(c *gin.Context, err error, defaultMsg string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		RespondError(c, http.StatusNotFound, "resource not found")
	case errors.Is(err, apperrors.ErrDuplicateKey):
		RespondError(c, http.StatusConflict, "resource already exists")
	case errors.Is(err, apperrors.ErrNoSeatsAvailable):
		RespondError(c, http.StatusConflict, "one or more selected seats are not available")
	case errors.Is(err, apperrors.ErrBookingNotCancellable):
		RespondError(c, http.StatusBadRequest, "booking cannot be cancelled")
	case errors.Is(err, domain.ErrArrivalBeforeDeparture):
		RespondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		RespondError(c, http.StatusBadRequest, err.Error())
	default:
		RespondError(c, http.StatusInternalServerError, defaultMsg)
	}
}
