package domain

import (
	"context"
	"time"
)

// UserRole represents permission levels in the application.
type UserRole string

const (
	UserRolePassenger UserRole = "passenger"
	UserRoleAdmin     UserRole = "admin"
)

// User is the domain entity for a registered user.
type User struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Password  string // Hashed password
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RegisterInput carries data required to create a new user account.
type RegisterInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
	Role     UserRole
}

// LoginInput carries user credentials for authentication.
type LoginInput struct {
	Email    string
	Password string
}

// AuthPayload contains the JWT token and public user details returned on login/register.
type AuthPayload struct {
	Token string
	User  *User
}

// UserRepository defines the data-access contract for users.
type UserRepository interface {
	Create(ctx context.Context, input RegisterInput, hashedPassword string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

// AuthService defines the business logic contract for authentication and user management.
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthPayload, error)
	Login(ctx context.Context, input LoginInput) (*AuthPayload, error)
	GetProfile(ctx context.Context, userID string) (*User, error)
	ValidateToken(tokenStr string) (userID, email, role string, err error)
}
