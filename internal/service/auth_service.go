package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
)

// JWTClaims defines custom payload stored inside the JWT token.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo    domain.UserRepository
	jwtSecret   []byte
	expiryHours time.Duration
}

// NewAuthService creates a new authService implementing domain.AuthService.
func NewAuthService(userRepo domain.UserRepository, jwtSecret string, expiryHours int) domain.AuthService {
	if expiryHours <= 0 {
		expiryHours = 24
	}
	return &authService{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		expiryHours: time.Duration(expiryHours) * time.Hour,
	}
}

// Register hashes user password, creates user record, and returns JWT token + User.
func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthPayload, error) {
	// Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, apperrors.ErrUserAlreadyExists
	}

	// Hash password using bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("auth_service: failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// Create user
	user, err := s.userRepo.Create(ctx, input, string(hashed))
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateKey) {
			return nil, apperrors.ErrUserAlreadyExists
		}
		slog.Error("auth_service: failed to create user", "email", input.Email, "error", err)
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		slog.Error("auth_service: failed to generate token", "user_id", user.ID, "error", err)
		return nil, err
	}

	slog.Info("user registered successfully", "id", user.ID, "email", user.Email, "role", user.Role)
	return &domain.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

// Login verifies email & password, and returns JWT token + User on success.
func (s *authService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthPayload, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		slog.Error("auth_service: error fetching user for login", "email", input.Email, "error", err)
		return nil, err
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		slog.Error("auth_service: failed to generate token on login", "user_id", user.ID, "error", err)
		return nil, err
	}

	slog.Info("user logged in successfully", "id", user.ID, "email", user.Email)
	return &domain.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

// GetProfile retrieves public user information by ID.
func (s *authService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("auth_service: error getting user profile", "user_id", userID, "error", err)
		return nil, err
	}
	return user, nil
}

// ValidateToken parses and validates a signed JWT string.
func (s *authService) ValidateToken(tokenStr string) (userID, email, role string, err error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", "", "", apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return "", "", "", apperrors.ErrInvalidToken
	}

	return claims.UserID, claims.Email, claims.Role, nil
}

// generateJWT creates a signed JWT string for a user.
func (s *authService) generateJWT(user *domain.User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiryHours)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
