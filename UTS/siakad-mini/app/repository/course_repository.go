package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) List(ctx context.Context, f model.CourseFilterQuery) ([]model.Course, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if f.Semester > 0 {
		conditions = append(conditions, fmt.Sprintf("c.semester = $%d", argIdx))
		args = append(args, f.Semester)
		argIdx++
	}

	if f.Search != "" {
		pattern := "%" + strings.TrimSpace(f.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	havingClause := ""
	if f.Available {
		havingClause = "HAVING (c.kuota - COUNT(e.id)) > 0"
	}

	query := fmt.Sprintf(`
		SELECT 
			c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota,
			c.created_at, c.updated_at
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		%s
		GROUP BY c.id
		%s
		ORDER BY c.id ASC
	`, whereClause, havingClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar mata kuliah: %w", err)
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(
			&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
			&c.Terisi, &c.SisaKuota, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal scan data mata kuliah: %w", err)
		}
		courses = append(courses, c)
	}

	return courses, nil
}

func (r *CourseRepository) FindByID(ctx context.Context, id int64) (*model.Course, error) {
	query := `
		SELECT 
			c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota,
			c.created_at, c.updated_at
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		WHERE c.id = $1
		GROUP BY c.id
	`
	row := r.pool.QueryRow(ctx, query, id)

	var c model.Course
	err := row.Scan(
		&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
		&c.Terisi, &c.SisaKuota, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari mata kuliah: %w", err)
	}

	return &c, nil
}

// FindByIDWithLock melakukan SELECT ... FOR UPDATE pada tabel courses dan menghitung jumlah terdaftar aktif
func (r *CourseRepository) FindByIDWithLock(ctx context.Context, tx pgx.Tx, id int64) (*model.Course, int, error) {
	query := `
		SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		FROM courses
		WHERE id = $1
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, id)

	var c model.Course
	err := row.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("gagal mengunci data mata kuliah: %w", err)
	}

	var terisi int
	countQuery := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	err = tx.QueryRow(ctx, countQuery, id).Scan(&terisi)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung kuota terisi: %w", err)
	}

	c.Terisi = terisi
	c.SisaKuota = c.Kuota - terisi

	return &c, terisi, nil
}
