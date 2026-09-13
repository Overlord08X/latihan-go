package repository

import (
	"context"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type pgUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	query := `
		INSERT INTO users (username, email, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		// Duplikat username atau email
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return u, nil
}

func (r *pgUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`
	u := &model.User{}
	err := r.db.QueryRow(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at
		FROM users
		WHERE id = $1
	`
	u := &model.User{}
	err := r.db.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}
