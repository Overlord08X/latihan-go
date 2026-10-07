package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"siakad-mini/app/model"
	"siakad-mini/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

func (r *StudentRepository) List(ctx context.Context, f model.StudentFilterQuery) ([]model.Student, *helper.MetaPagination, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	perPage := f.PerPage
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	var conditions []string
	var args []any
	argIdx := 1

	// Wajib soft delete check
	conditions = append(conditions, "deleted_at IS NULL")

	if f.Prodi != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(prodi) = LOWER($%d)", argIdx))
		args = append(args, f.Prodi)
		argIdx++
	}

	if f.Angkatan > 0 {
		conditions = append(conditions, fmt.Sprintf("angkatan = $%d", argIdx))
		args = append(args, f.Angkatan)
		argIdx++
	}

	if f.Search != "" {
		pattern := "%" + strings.TrimSpace(f.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(nama ILIKE $%d OR nim ILIKE $%d)", argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Hitung total data
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal menghitung total data mahasiswa: %w", err)
	}

	lastPage := 1
	if total > 0 {
		lastPage = (total + perPage - 1) / perPage
	}

	// Sorting
	orderClause := "ORDER BY id ASC"
	switch f.Sort {
	case "nama":
		orderClause = "ORDER BY nama ASC"
	case "-nama":
		orderClause = "ORDER BY nama DESC"
	case "ipk_terakhir":
		orderClause = "ORDER BY ipk_terakhir ASC, id ASC"
	case "-ipk_terakhir":
		orderClause = "ORDER BY ipk_terakhir DESC, id ASC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		FROM students
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIdx, argIdx+1)

	queryArgs := append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("gagal scan data mahasiswa: %w", err)
		}
		students = append(students, s)
	}

	meta := &helper.MetaPagination{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return students, meta, nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id int64) (*model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		FROM students
		WHERE id = $1 AND deleted_at IS NULL
	`
	row := r.pool.QueryRow(ctx, query, id)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari mahasiswa berdasarkan id: %w", err)
	}

	return &s, nil
}

func (r *StudentRepository) FindByUserID(ctx context.Context, userID int64) (*model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		FROM students
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	row := r.pool.QueryRow(ctx, query, userID)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari mahasiswa berdasarkan user_id: %w", err)
	}

	return &s, nil
}

func (r *StudentRepository) ExistsByNIM(ctx context.Context, nim string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE nim = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nim).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("gagal cek eksistensi nim: %w", err)
	}
	return exists, nil
}

func (r *StudentRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("gagal cek eksistensi email: %w", err)
	}
	return exists, nil
}

func (r *StudentRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, s *model.Student) (int64, error) {
	query := `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	var id int64
	err := tx.QueryRow(ctx, query, s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir).Scan(&id, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("gagal membuat record mahasiswa: %w", err)
	}
	s.ID = id
	return id, nil
}

func (r *StudentRepository) Update(ctx context.Context, id int64, nama, prodi string, angkatan int, ipk float64) (*model.Student, error) {
	query := `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query, nama, prodi, angkatan, ipk, id)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengupdate mahasiswa: %w", err)
	}

	return &s, nil
}

func (r *StudentRepository) SoftDelete(ctx context.Context, id int64) (bool, error) {
	query := `
		UPDATE students
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("gagal soft delete mahasiswa: %w", err)
	}
	return cmd.RowsAffected() > 0, nil
}
