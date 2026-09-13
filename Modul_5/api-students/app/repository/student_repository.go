// Package repository menyediakan lapisan akses data (data access layer).
// Pola Repository memisahkan logika bisnis dari detail implementasi database,
// sehingga handler tidak perlu tahu apakah data disimpan di PostgreSQL, MongoDB,
// atau bahkan in-memory — cukup memanggil method pada interface.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
// Sentinel errors — error yang bisa dibandingkan dengan errors.Is()
// -----------------------------------------------------------------------------

// ErrNotFound dikembalikan ketika data yang dicari tidak ada di database.
var ErrNotFound = errors.New("data tidak ditemukan")

// ErrDuplicate dikembalikan ketika terjadi pelanggaran constraint unique
// (misalnya NIM yang sama).
var ErrDuplicate = errors.New("data sudah ada (konflik)")

// -----------------------------------------------------------------------------
// Interface — kontrak yang harus dipenuhi implementasi apa pun
// -----------------------------------------------------------------------------

// ListParams menyimpan parameter untuk query daftar mahasiswa.
type ListParams struct {
	Page     int
	Limit    int
	Search   string
	SortBy   string
	Order    string
	IsActive *bool
}

// StudentRepository mendefinisikan operasi CRUD yang tersedia untuk entitas Student.
// Dengan menggunakan interface, handler bisa diuji tanpa memerlukan koneksi database
// nyata (cukup buat mock yang mengimplementasikan interface ini).
type StudentRepository interface {
	List(ctx context.Context, p ListParams) ([]*model.Student, int, error)
	GetByID(ctx context.Context, id int) (*model.Student, error)
	Create(ctx context.Context, s *model.Student) (*model.Student, error)
	Replace(ctx context.Context, id int, s *model.Student) (*model.Student, error)
	Patch(ctx context.Context, id int, fields map[string]interface{}) (*model.Student, error)
	Delete(ctx context.Context, id int) error
}

// -----------------------------------------------------------------------------
// Implementasi PostgreSQL
// -----------------------------------------------------------------------------

// pgStudentRepository adalah implementasi StudentRepository yang menggunakan pgxpool.
type pgStudentRepository struct {
	db *pgxpool.Pool
}

// NewStudentRepository membuat instance pgStudentRepository baru.
// Handler menerima interface StudentRepository, bukan struct konkret,
// sehingga implementasi bisa diganti tanpa mengubah handler.
func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &pgStudentRepository{db: db}
}

// terjemahkanError mengubah error pgx menjadi sentinel error yang dapat
// diinterpretasikan oleh handler untuk menentukan status HTTP yang tepat.
func terjemahkanError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	// Kode "23505" adalah SQLSTATE untuk unique_violation di PostgreSQL.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return err
}

// List mengambil daftar mahasiswa dengan filter, sort, dan paginasi.
func (r *pgStudentRepository) List(ctx context.Context, p ListParams) ([]*model.Student, int, error) {
	// Daftar putih kolom yang boleh diurutkan untuk mencegah SQL injection
	allowedSort := map[string]string{
		"id": "id", "nim": "nim", "name": "name", "grade": "grade",
	}
	sortCol, ok := allowedSort[p.SortBy]
	if !ok {
		sortCol = "id"
	}
	orderDir := "ASC"
	if strings.ToLower(p.Order) == "desc" {
		orderDir = "DESC"
	}

	// Bangun WHERE clause secara dinamis
	args := []interface{}{}
	wheres := []string{}
	argIdx := 1

	if p.Search != "" {
		wheres = append(wheres, fmt.Sprintf("LOWER(name) LIKE $%d", argIdx))
		args = append(args, "%"+strings.ToLower(p.Search)+"%")
		argIdx++
	}
	if p.IsActive != nil {
		wheres = append(wheres, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *p.IsActive)
		argIdx++
	}

	whereClause := ""
	if len(wheres) > 0 {
		whereClause = "WHERE " + strings.Join(wheres, " AND ")
	}

	// Hitung total sebelum paginasi
	countSQL := "SELECT COUNT(*) FROM students " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Query data dengan paginasi
	offset := (p.Page - 1) * p.Limit
	dataSQL := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students %s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, orderDir, argIdx, argIdx+1,
	)
	args = append(args, p.Limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []*model.Student{}
	for rows.Next() {
		s := &model.Student{}
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetByID mengambil satu mahasiswa berdasarkan ID.
// Mengembalikan ErrNotFound jika tidak ada.
func (r *pgStudentRepository) GetByID(ctx context.Context, id int) (*model.Student, error) {
	s := &model.Student{}
	err := r.db.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)
	if err != nil {
		return nil, terjemahkanError(err)
	}
	return s, nil
}

// Create menyimpan mahasiswa baru dan mengembalikan data lengkapnya (termasuk ID
// dan created_at yang di-generate server).
func (r *pgStudentRepository) Create(ctx context.Context, s *model.Student) (*model.Student, error) {
	result := &model.Student{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive,
	).Scan(&result.ID, &result.NIM, &result.Name, &result.Grade, &result.IsActive, &result.CreatedAt)
	if err != nil {
		return nil, terjemahkanError(err)
	}
	return result, nil
}

// Replace mengganti seluruh data mahasiswa (PUT). Idempoten: hasil akhir sama
// tidak peduli berapa kali dijalankan.
func (r *pgStudentRepository) Replace(ctx context.Context, id int, s *model.Student) (*model.Student, error) {
	result := &model.Student{}
	err := r.db.QueryRow(ctx,
		`UPDATE students
		 SET nim = $1, name = $2, grade = $3, is_active = $4
		 WHERE id = $5
		 RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, id,
	).Scan(&result.ID, &result.NIM, &result.Name, &result.Grade, &result.IsActive, &result.CreatedAt)
	if err != nil {
		return nil, terjemahkanError(err)
	}
	return result, nil
}

// Patch mengubah hanya field yang ada di map fields (PATCH).
// Teknik ini menghindari overwrite field yang tidak dikirim klien.
func (r *pgStudentRepository) Patch(ctx context.Context, id int, fields map[string]interface{}) (*model.Student, error) {
	// Kolom yang boleh di-patch (whitelist)
	allowed := map[string]bool{"nim": true, "name": true, "grade": true, "is_active": true}

	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	for col, val := range fields {
		if !allowed[col] {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if len(sets) == 0 {
		// Tidak ada field yang diubah; kembalikan data yang ada
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	sql := fmt.Sprintf(
		`UPDATE students SET %s WHERE id = $%d
		 RETURNING id, nim, name, grade, is_active, created_at`,
		strings.Join(sets, ", "), argIdx,
	)

	result := &model.Student{}
	err := r.db.QueryRow(ctx, sql, args...).Scan(
		&result.ID, &result.NIM, &result.Name, &result.Grade, &result.IsActive, &result.CreatedAt,
	)
	if err != nil {
		return nil, terjemahkanError(err)
	}
	return result, nil
}

// Delete menghapus mahasiswa berdasarkan ID.
// Mengembalikan ErrNotFound jika tidak ada baris yang terhapus.
func (r *pgStudentRepository) Delete(ctx context.Context, id int) error {
	ct, err := r.db.Exec(ctx, "DELETE FROM students WHERE id = $1", id)
	if err != nil {
		return terjemahkanError(err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
