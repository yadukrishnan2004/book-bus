package models

import (
	"time"

	"book-bus/internal/domain"
)

// RegisterRequest is the JSON payload for user registration.
// The role field is intentionally absent — public registration always creates
// a "passenger" account. Admin accounts must be seeded directly in the DB.
type RegisterRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=255"`
	Email    string `json:"email"    binding:"required,email,max=255"`
	Phone    string `json:"phone"    binding:"required,min=7,max=20"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// ToDomainInput converts RegisterRequest to domain.RegisterInput.
// Role is always set to passenger — callers cannot self-assign admin.
func (r RegisterRequest) ToDomainInput() domain.RegisterInput {
	return domain.RegisterInput{
		Name:     r.Name,
		Email:    r.Email,
		Phone:    r.Phone,
		Password: r.Password,
		Role:     domain.UserRolePassenger,
	}
}

// LoginRequest is the JSON payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// ToDomainInput converts LoginRequest to domain.LoginInput.
func (r LoginRequest) ToDomainInput() domain.LoginInput {
	return domain.LoginInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

// UserResponse is the sanitized user representation returned in API responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserResponse maps a domain.User to UserResponse.
func NewUserResponse(u *domain.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// AuthResponse is the JSON payload returned upon successful login or registration.
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// NewAuthResponse maps domain.AuthPayload to AuthResponse.
func NewAuthResponse(p *domain.AuthPayload) *AuthResponse {
	if p == nil {
		return nil
	}
	return &AuthResponse{
		Token: p.Token,
		User:  NewUserResponse(p.User),
	}
}
