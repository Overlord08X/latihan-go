package repository

import (
	"context"
	"errors"
	"fmt"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password, role, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	row := r.pool.QueryRow(ctx, query, email)

	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari user berdasarkan email: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari user berdasarkan id: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, email, password, role string) (int64, error) {
	query := `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	err := tx.QueryRow(ctx, query, email, password, role).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("gagal membuat user: %w", err)
	}
	return id, nil
}
