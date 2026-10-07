package service

import (
	"context"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type StudentService struct {
	pool           *pgxpool.Pool
	studentRepo    *repository.StudentRepository
	userRepo       *repository.UserRepository
	enrollmentRepo *repository.EnrollmentRepository
}

func NewStudentService(
	pool *pgxpool.Pool,
	studentRepo *repository.StudentRepository,
	userRepo *repository.UserRepository,
	enrollmentRepo *repository.EnrollmentRepository,
) *StudentService {
	return &StudentService{
		pool:           pool,
		studentRepo:    studentRepo,
		userRepo:       userRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

func (s *StudentService) List(ctx context.Context, f model.StudentFilterQuery) ([]model.Student, *helper.MetaPagination, error) {
	students, meta, err := s.studentRepo.List(ctx, f)
	if err != nil {
		return nil, nil, helper.Internal(err, "Gagal mengambil daftar mahasiswa")
	}
	if students == nil {
		students = []model.Student{}
	}
	return students, meta, nil
}

func (s *StudentService) Create(ctx context.Context, req model.CreateStudentRequest) (*model.Student, error) {
	// 1. Validasi deklaratif
	v := helper.GetValidator()
	if err := v.Struct(req); err != nil {
		return nil, helper.UnprocessableEntity("Validasi gagal", helper.FormatValidationErrors(err))
	}

	// 2. Cek duplikasi NIM dan Email sebelum transaksi
	nimExists, err := s.studentRepo.ExistsByNIM(ctx, req.NIM)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memeriksa duplikasi NIM")
	}
	if nimExists {
		return nil, helper.UnprocessableEntity("Validasi gagal", map[string][]string{
			"nim": {"NIM sudah terdaftar dalam sistem"},
		})
	}

	emailExists, err := s.studentRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memeriksa duplikasi email")
	}
	if emailExists {
		return nil, helper.UnprocessableEntity("Validasi gagal", map[string][]string{
			"email": {"Email sudah terdaftar dalam sistem"},
		})
	}

	// 3. Hash default password = NIM
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		return nil, helper.Internal(err, "Gagal meng-hash password awal mahasiswa")
	}

	// 4. Jalankan transaksi pembuatan akun user dan mahasiswa secara atomik
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memulai transaksi database")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	userID, err := s.userRepo.CreateWithTx(ctx, tx, req.Email, string(hashedPassword), "mahasiswa")
	if err != nil {
		return nil, helper.Internal(err, "Gagal menyimpan akun user mahasiswa")
	}

	ipk := 0.00
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	newStudent := &model.Student{
		UserID:      userID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}

	_, err = s.studentRepo.CreateWithTx(ctx, tx, newStudent)
	if err != nil {
		return nil, helper.Internal(err, "Gagal menyimpan record mahasiswa")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, helper.Internal(err, "Gagal menyelesaikan transaksi pendaftaran mahasiswa")
	}

	return newStudent, nil
}

func (s *StudentService) Get(ctx context.Context, authRole string, authStudentID, targetID int64) (*model.StudentDetail, error) {
	// 1. Kontrol Akses: Mahasiswa hanya boleh melihat datanya sendiri
	if !helper.CanAccessStudent(authRole, authStudentID, targetID) {
		return nil, helper.Forbidden("Akses ditolak: mahasiswa hanya dapat mengakses profil dan KRS miliknya sendiri")
	}

	// 2. Ambil data mahasiswa
	student, err := s.studentRepo.FindByID(ctx, targetID)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil data mahasiswa")
	}
	if student == nil {
		return nil, helper.NotFound("Mahasiswa tidak ditemukan atau telah dinonaktifkan")
	}

	// 3. Ambil daftar mata kuliah yang diambil
	courses, totalSKS, err := s.enrollmentRepo.ListCoursesByStudentID(ctx, student.ID)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil data KRS mahasiswa")
	}
	if courses == nil {
		courses = []model.EnrolledCourse{}
	}

	// 4. Hitung batas SKS berdasarkan aturan IPK
	batasSKS := helper.GetMaxSKS(student.IPKTerakhir)

	return &model.StudentDetail{
		Student:     *student,
		MataKuliah:  courses,
		TotalSKS:    totalSKS,
		BatasSKS:    batasSKS,
	}, nil
}

func (s *StudentService) Update(ctx context.Context, id int64, req model.UpdateStudentRequest) (*model.Student, error) {
	// 1. Validasi deklaratif
	v := helper.GetValidator()
	if err := v.Struct(req); err != nil {
		return nil, helper.UnprocessableEntity("Validasi gagal", helper.FormatValidationErrors(err))
	}

	// 2. Cek keberadaan mahasiswa
	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memeriksa mahasiswa")
	}
	if existing == nil {
		return nil, helper.NotFound("Mahasiswa tidak ditemukan")
	}

	ipk := existing.IPKTerakhir
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	// 3. Eksekusi pembaruan data
	updated, err := s.studentRepo.Update(ctx, id, req.Nama, req.Prodi, req.Angkatan, ipk)
	if err != nil {
		return nil, helper.Internal(err, "Gagal memperbarui data mahasiswa")
	}
	if updated == nil {
		return nil, helper.NotFound("Mahasiswa tidak ditemukan")
	}

	return updated, nil
}

func (s *StudentService) Delete(ctx context.Context, id int64) error {
	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err, "Gagal memeriksa mahasiswa")
	}
	if existing == nil {
		return helper.NotFound("Mahasiswa tidak ditemukan atau sudah dihapus sebelumnya")
	}

	success, err := s.studentRepo.SoftDelete(ctx, id)
	if err != nil {
		return helper.Internal(err, "Gagal menghapus mahasiswa")
	}
	if !success {
		return helper.NotFound("Mahasiswa tidak ditemukan")
	}

	return nil
}
