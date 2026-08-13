package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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

// RespondValidationError handles binding/validation errors and formats them cleanly into JSON.
func RespondValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string, len(ve))
		for _, fe := range ve {
			jsonTag := fe.Field()
			// Convert struct field to lower case json field if needed
			field := strings.ToLower(jsonTag)
			out[field] = formatFieldError(fe)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": out,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "invalid request format",
		"details": err.Error(),
	})
}

// RespondBadRequest sends a 400 Bad Request error response with details.
func RespondBadRequest(c *gin.Context, message string, details string) {
	if details != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": message, "details": details})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// HandleError maps domain & app errors to standardized HTTP responses.
func HandleError(c *gin.Context, err error, defaultMsg string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		RespondError(c, http.StatusNotFound, "resource not found")
	case errors.Is(err, apperrors.ErrDuplicateKey), errors.Is(err, apperrors.ErrUserAlreadyExists):
		RespondError(c, http.StatusConflict, err.Error())
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		RespondError(c, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, apperrors.ErrUnauthorized), errors.Is(err, apperrors.ErrInvalidToken):
		RespondError(c, http.StatusUnauthorized, "unauthorized access")
	case errors.Is(err, apperrors.ErrForbidden):
		RespondError(c, http.StatusForbidden, "forbidden")
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

func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters/items", fe.Param())
	case "max":
		return fmt.Sprintf("must not exceed %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation on tag '%s'", fe.Tag())
	}
}
