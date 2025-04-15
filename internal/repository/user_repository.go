package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/auliayudha/oapi-sample/internal/gen"
)

var (
	ErrNotFound = errors.New("user not found")
)

// UserRepository manages database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// List returns all users
func (r *UserRepository) List(ctx context.Context) ([]gen.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []gen.User
	for rows.Next() {
		var user gen.User
		if err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Get returns a user by ID
func (r *UserRepository) Get(ctx context.Context, id int64) (gen.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	var user gen.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return gen.User{}, ErrNotFound
		}
		return gen.User{}, err
	}

	return user, nil
}

// Create adds a new user
func (r *UserRepository) Create(ctx context.Context, req gen.UserRequest) (gen.User, error) {
	query := `
		INSERT INTO users (name, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, created_at, updated_at
	`

	now := time.Now().UTC()
	var user gen.User
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Email, now, now).Scan(
		&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return gen.User{}, err
	}

	return user, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, id int64, req gen.UserRequest) (gen.User, error) {
	query := `
		UPDATE users 
		SET name = $1, email = $2, updated_at = $3 
		WHERE id = $4 
		RETURNING id, name, email, created_at, updated_at
	`

	now := time.Now().UTC()
	var user gen.User
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Email, now, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return gen.User{}, ErrNotFound
		}
		return gen.User{}, err
	}

	return user, nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
