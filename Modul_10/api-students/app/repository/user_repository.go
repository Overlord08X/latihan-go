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

const userColumns = "id, username, email, password, role, is_active, created_at"

type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.User, error)
	Update(ctx context.Context, id int, u model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

type pgUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgUserRepository{db: db}
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (model.User, error) {
	var u model.User
	err := scanner.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
	)
	return u, err
}

func (r *pgUserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	query := `
		INSERT INTO users (username, email, password, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role, true).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	u.IsActive = true
	return u, nil
}

func (r *pgUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE LOWER(username) = LOWER($1)`, userColumns)
	u, err := scanUser(r.db.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	return r.FindByID(ctx, id)
}

func (r *pgUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)
	u, err := scanUser(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepository) List(ctx context.Context) ([]model.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users ORDER BY created_at DESC, id DESC`, userColumns)
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("mengambil list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// FindAfterCursor mengambil satu halaman memakai keyset pagination (Langkah 7).
// Perbaikan Bug Perilaku #8: Menggunakan ORDER BY created_at DESC, id DESC
// agar halaman terurut dari yang terbaru ke terlama dan perbandingan row value < ($1, $2) berjalan benar.
func (r *pgUserRepository) FindAfterCursor(
	ctx context.Context, q model.CursorQuery,
) ([]model.User, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND username ILIKE $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)",
			len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"SELECT %s FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d",
		userColumns, where, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	result := []model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row user: %w", err)
		}
		result = append(result, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}
	return result, nil
}

func (r *pgUserRepository) Update(ctx context.Context, id int, u model.User) (model.User, error) {
	query := fmt.Sprintf(`
		UPDATE users
		SET username = $1, email = $2, is_active = $3
		WHERE id = $4
		RETURNING %s
	`, userColumns)
	res, err := scanUser(r.db.QueryRow(ctx, query, u.Username, u.Email, u.IsActive, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "no rows in result set" {
			return model.User{}, ErrNotFound
		}
		if strings.Contains(err.Error(), "unique") {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("mengupdate user: %w", err)
	}
	return res, nil
}

func (r *pgUserRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	query := fmt.Sprintf(`
		UPDATE users
		SET role = $1
		WHERE id = $2
		RETURNING %s
	`, userColumns)
	u, err := scanUser(r.db.QueryRow(ctx, query, role, id))
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
