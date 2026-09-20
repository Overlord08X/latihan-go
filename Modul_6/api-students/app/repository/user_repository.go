package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
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
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	return r.FindByID(ctx, id)
}

func (r *pgUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at
		FROM users
		WHERE id = $1
	`
	u := &model.User{}
	err := r.db.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *pgUserRepository) List(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at
		FROM users
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("mengambil list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *pgUserRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	query := `
		UPDATE users
		SET role = $1
		WHERE id = $2
		RETURNING id, username, email, password, role, created_at
	`
	var u model.User
	err := r.db.QueryRow(ctx, query, role, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return u, nil
}

func (r *pgUserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("menghapus user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
