package service

import (
	"context"
	"fmt"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EnrollmentService struct {
	pool           *pgxpool.Pool
	enrollmentRepo *repository.EnrollmentRepository
	courseRepo     *repository.CourseRepository
	studentRepo    *repository.StudentRepository
}

func NewEnrollmentService(
	pool *pgxpool.Pool,
	enrollmentRepo *repository.EnrollmentRepository,
	courseRepo *repository.CourseRepository,
	studentRepo *repository.StudentRepository,
) *EnrollmentService {
	return &EnrollmentService{
		pool:           pool,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		studentRepo:    studentRepo,
	}
}

type EnrollmentCreatedResponse struct {
	ID            int64  `json:"id"`
	StudentID     int64  `json:"student_id"`
	CourseID      int64  `json:"course_id"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	TahunAkademik string `json:"tahun_akademik"`
}

func (s *EnrollmentService) Create(ctx context.Context, userID int64, req model.CreateEnrollmentRequest) (*EnrollmentCreatedResponse, error) {
	// 1. Validasi deklaratif body
	v := helper.GetValidator()
	if err := v.Struct(req); err != nil {
		return nil, helper.UnprocessableEntity("Validasi gagal", helper.FormatValidationErrors(err))
	}

	// 2. Ambil data mahasiswa pemilik akun login
	student, err := s.studentRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil identitas mahasiswa")
	}
	if student == nil {
		return nil, helper.Forbidden("Akses ditolak: akun tidak terdaftar sebagai mahasiswa aktif")
	}

	// 3. Mulai transaksi dengan row locking
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memulai transaksi pendaftaran KRS")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 4. Aturan Bisnis 2: Cek pengambilan ganda pada tahun akademik yang sama
	alreadyEnrolled, err := s.enrollmentRepo.ExistsWithTx(ctx, tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memeriksa riwayat KRS")
	}
	if alreadyEnrolled {
		return nil, helper.Conflict("Mata kuliah ini sudah diambil pada tahun akademik yang sama")
	}

	// 5. Aturan Bisnis 3: Kunci baris mata kuliah (SELECT ... FOR UPDATE) dan cek kuota
	course, _, err := s.courseRepo.FindByIDWithLock(ctx, tx, req.CourseID)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memeriksa kuota mata kuliah")
	}
	if course == nil {
		return nil, helper.NotFound("Mata kuliah tidak ditemukan")
	}

	if course.Terisi >= course.Kuota {
		return nil, helper.NewError(422, fmt.Sprintf("Kuota mata kuliah '%s' sudah penuh (kuota: %d, terisi: %d)", course.NamaMK, course.Kuota, course.Terisi))
	}

	// 6. Aturan Bisnis 1: Evaluasi batas SKS berdasarkan IPK terakhir
	batasSKS := helper.GetMaxSKS(student.IPKTerakhir)

	currentSKS, err := s.enrollmentRepo.GetCurrentSKSWithTx(ctx, tx, student.ID, req.TahunAkademik)
	if err != nil {
		return nil, helper.Internal(err, "Gagal menghitung akumulasi SKS semester")
	}

	if currentSKS+course.SKS > batasSKS {
		sisaSKS := batasSKS - currentSKS
		msg := fmt.Sprintf("Batas SKS terlampaui. Batas maksimal Anda adalah %d SKS, sudah diambil %d SKS, sisa SKS Anda adalah %d SKS, sedangkan mata kuliah ini berbobot %d SKS", batasSKS, currentSKS, sisaSKS, course.SKS)
		return nil, helper.NewError(422, msg)
	}

	// 7. Simpan enrollment baru
	enrollmentID, err := s.enrollmentRepo.CreateWithTx(ctx, tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return nil, helper.Internal(err, "Gagal menyimpan entri KRS")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, helper.Internal(err, "Gagal menyelesaikan transaksi KRS")
	}

	return &EnrollmentCreatedResponse{
		ID:            enrollmentID,
		StudentID:     student.ID,
		CourseID:      course.ID,
		NamaMK:        course.NamaMK,
		SKS:           course.SKS,
		TahunAkademik: req.TahunAkademik,
	}, nil
}

func (s *EnrollmentService) Delete(ctx context.Context, userID int64, enrollmentID int64) error {
	// 1. Ambil identitas mahasiswa yang sedang login
	student, err := s.studentRepo.FindByUserID(ctx, userID)
	if err != nil {
		return helper.Internal(err, "Gagal mengambil identitas mahasiswa")
	}
	if student == nil {
		return helper.Forbidden("Akses ditolak: akun tidak terdaftar sebagai mahasiswa aktif")
	}

	// 2. Ambil data enrollment yang ingin dibatalkan
	enrollment, err := s.enrollmentRepo.FindByID(ctx, enrollmentID)
	if err != nil {
		return helper.Internal(err, "Gagal mencari data enrollment")
	}
	if enrollment == nil {
		return helper.NotFound("Data KRS tidak ditemukan")
	}

	// 3. Aturan Bisnis 4: Mahasiswa hanya dapat membatalkan KRS miliknya sendiri
	if enrollment.StudentID != student.ID {
		return helper.Forbidden("Anda tidak berhak membatalkan KRS mahasiswa lain")
	}

	// 4. Hapus enrollment (otomatis mengembalikan kuota mata kuliah)
	if err := s.enrollmentRepo.Delete(ctx, enrollmentID); err != nil {
		return helper.Internal(err, "Gagal membatalkan mata kuliah dari KRS")
	}

	return nil
}
