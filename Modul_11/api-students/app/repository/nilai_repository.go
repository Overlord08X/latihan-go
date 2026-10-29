package repository

import (
	"context"

	"api-students/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NilaiRepository interface {
	Create(ctx context.Context, n *model.Nilai) (*model.Nilai, error)
	GetByStudentNIM(ctx context.Context, nim string) ([]*model.Nilai, error)
}

type pgNilaiRepository struct {
	db *pgxpool.Pool
}

func NewNilaiRepository(db *pgxpool.Pool) NilaiRepository {
	return &pgNilaiRepository{db: db}
}

func (r *pgNilaiRepository) Create(ctx context.Context, n *model.Nilai) (*model.Nilai, error) {
	query := `
		INSERT INTO nilais (id_student, nama_matkul, nilai)
		VALUES ($1, $2, $3)
		RETURNING id_nilai, created_at
	`
	err := r.db.QueryRow(ctx, query, n.IDStudent, n.NamaMatkul, n.Nilai).Scan(&n.IDNilai, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *pgNilaiRepository) GetByStudentNIM(ctx context.Context, nim string) ([]*model.Nilai, error) {
	query := `
		SELECT n.id_nilai, n.id_student, n.nama_matkul, n.nilai, n.created_at
		FROM nilais n
		JOIN students s ON n.id_student = s.id
		WHERE s.nim = $1
		ORDER BY n.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nilais []*model.Nilai
	for rows.Next() {
		n := &model.Nilai{}
		err := rows.Scan(
			&n.IDNilai,
			&n.IDStudent,
			&n.NamaMatkul,
			&n.Nilai,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nilais = append(nilais, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Cek apakah mahasiswa exist berdasarkan NIM?
	// Kalau daftar nilai kosong, bisa saja karena mahasiswanya tidak ada, atau memang belum ada nilainya.
	// Tetapi berdasarkan spesifikasi endpoint ini cukup kembalikan list kosong jika tak ada nilai.
	// Jika ingin ketat, kita periksa apakah NIM-nya ada.
	if len(nilais) == 0 {
		var exists bool
		err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM students WHERE nim=$1)", nim).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}

	return nilais, nil
}
