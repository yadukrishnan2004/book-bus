package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/domain"
)

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new userRepository implementing domain.UserRepository.
func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user record into PostgreSQL.
func (r *userRepository) Create(ctx context.Context, input domain.RegisterInput, hashedPassword string) (*domain.User, error) {
	query := `
		INSERT INTO users (name, email, phone, password, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, phone, password, role, created_at, updated_at
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query,
		input.Name,
		input.Email,
		input.Phone,
		hashedPassword,
		input.Role,
	).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Phone,
		&u.Password,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "create user")
	}
	return u, nil
}

// GetByEmail retrieves a single user by their email address.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, phone, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Phone,
		&u.Password,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "get user by email")
	}
	return u, nil
}

// GetByID retrieves a single user by their UUID.
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, phone, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Phone,
		&u.Password,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "get user by id")
	}
	return u, nil
}
