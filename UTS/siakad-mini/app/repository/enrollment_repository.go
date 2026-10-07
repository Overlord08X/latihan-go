package repository

import (
	"context"
	"errors"
	"fmt"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (int64, error) {
	query := `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	err := tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("gagal menambahkan enrollment: %w", err)
	}
	return id, nil
}

func (r *EnrollmentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus enrollment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("enrollment tidak ditemukan")
	}
	return nil
}

func (r *EnrollmentRepository) FindByID(ctx context.Context, id int64) (*model.Enrollment, error) {
	query := `
		SELECT id, student_id, course_id, tahun_akademik, created_at
		FROM enrollments
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var e model.Enrollment
	err := row.Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari enrollment: %w", err)
	}

	return &e, nil
}

func (r *EnrollmentRepository) ExistsWithTx(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM enrollments 
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)
	`
	var exists bool
	err := tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa duplikasi enrollment: %w", err)
	}
	return exists, nil
}

func (r *EnrollmentRepository) GetCurrentSKSWithTx(ctx context.Context, tx pgx.Tx, studentID int64, tahunAkademik string) (int, error) {
	query := `
		SELECT COALESCE(SUM(c.sks), 0)::int
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	var total int
	err := tx.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung total SKS terdaftar: %w", err)
	}
	return total, nil
}

func (r *EnrollmentRepository) ListCoursesByStudentID(ctx context.Context, studentID int64) ([]model.EnrolledCourse, int, error) {
	query := `
		SELECT 
			e.id AS enrollment_id,
			c.id AS course_id,
			c.kode_mk,
			c.nama_mk,
			c.sks,
			c.semester,
			e.tahun_akademik
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil mata kuliah terdaftar: %w", err)
	}
	defer rows.Close()

	var list []model.EnrolledCourse
	totalSKS := 0
	for rows.Next() {
		var item model.EnrolledCourse
		if err := rows.Scan(
			&item.EnrollmentID, &item.CourseID, &item.KodeMK, &item.NamaMK,
			&item.SKS, &item.Semester, &item.TahunAkademik,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal scan enrolled course: %w", err)
		}
		totalSKS += item.SKS
		list = append(list, item)
	}

	return list, totalSKS, nil
}
