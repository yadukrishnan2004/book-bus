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

// AuthHandler handles HTTP endpoints for user registration, login, and profile.
type AuthHandler struct {
	authSvc domain.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc domain.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// RegisterRoutes registers auth endpoints onto the given router group.
func (h *AuthHandler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	authPublic := public.Group("/auth")
	{
		authPublic.POST("/register", h.Register)
		authPublic.POST("/login", h.Login)
	}

	authProtected := protected.Group("/auth")
	{
		authProtected.GET("/me", h.GetProfile)
	}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	payload, err := h.authSvc.Register(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			RespondError(c, http.StatusConflict, "user with this email already exists")
			return
		}
		HandleError(c, err, "registration failed")
		return
	}

	RespondJSON(c, http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"data":    models.NewAuthResponse(payload),
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}

	payload, err := h.authSvc.Login(c.Request.Context(), req.ToDomainInput())
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			RespondError(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		HandleError(c, err, "login failed")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"message": "login successful",
		"data":    models.NewAuthResponse(payload),
	})
}

// GetProfile handles GET /api/v1/auth/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		RespondError(c, http.StatusUnauthorized, "unauthorized access")
		return
	}

	user, err := h.authSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "user profile not found")
			return
		}
		HandleError(c, err, "failed to retrieve profile")
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"data": models.NewUserResponse(user),
	})
}
