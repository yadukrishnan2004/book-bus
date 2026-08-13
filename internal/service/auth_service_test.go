package service_test

import (
	"testing"

	"book-bus/internal/service"
)

func TestAuthService_JWTValidation(t *testing.T) {
	secret := "test-secret-key-12345"
	expiryHours := 1

	authSvc := service.NewAuthService(nil, secret, expiryHours)

	t.Run("Invalid Token Format", func(t *testing.T) {
		_, _, _, err := authSvc.ValidateToken("invalid.jwt.token")
		if err == nil {
			t.Errorf("expected error for invalid token, got nil")
		}
	})
}
