package service

import (
	"context"
	"fmt"
	"strings"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
	"siakad-mini/jwtutil"
	"siakad-mini/middleware"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	studentRepo    *repository.StudentRepository
	jwtSecret      string
	jwtExpHours    int
	failureTracker *middleware.FailureTracker
}

func NewAuthService(
	userRepo *repository.UserRepository,
	studentRepo *repository.StudentRepository,
	jwtSecret string,
	jwtExpHours int,
	failureTracker *middleware.FailureTracker,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		studentRepo:    studentRepo,
		jwtSecret:      jwtSecret,
		jwtExpHours:    jwtExpHours,
		failureTracker: failureTracker,
	}
}

type LoginResponseData struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        UserInfoDTO `json:"user"`
}

type UserInfoDTO struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type MeMahasiswaResponse struct {
	UserInfoDTO
	Student *model.Student `json:"student"`
}

func (s *AuthService) Login(ctx context.Context, clientIP string, req model.LoginRequest) (*LoginResponseData, error) {
	rateKey := fmt.Sprintf("%s:%s", clientIP, strings.ToLower(req.Email))

	// 1. Periksa batas laju kegagalan (Rate Limiting: maks 5 kegagalan per menit)
	if s.failureTracker.IsBlocked(rateKey) {
		return nil, helper.TooManyRequests("Terlalu banyak percobaan login gagal, silakan tunggu 1 menit sebelum mencoba kembali")
	}

	// 2. Validasi deklaratif body
	v := helper.GetValidator()
	if err := v.Struct(req); err != nil {
		return nil, helper.UnprocessableEntity("Validasi gagal", helper.FormatValidationErrors(err))
	}

	// 3. Cari pengguna berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil data user")
	}
	if user == nil {
		s.failureTracker.RecordFailure(rateKey)
		return nil, helper.Unauthorized("Kredensial login tidak valid")
	}

	// 4. Jika role mahasiswa, pastikan tidak dalam status soft-deleted
	var studentID int64
	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, helper.Internal(err, "Gagal memverifikasi akun mahasiswa")
		}
		if student == nil {
			s.failureTracker.RecordFailure(rateKey)
			return nil, helper.Unauthorized("Akun mahasiswa telah dinonaktifkan atau dihapus")
		}
		studentID = student.ID
	}

	// 5. Bandingkan password bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.failureTracker.RecordFailure(rateKey)
		return nil, helper.Unauthorized("Kredensial login tidak valid")
	}

	// 6. Login berhasil -> reset counter kegagalan
	s.failureTracker.Reset(rateKey)

	// 7. Terbitkan Access Token JWT
	token, exp, err := jwtutil.GenerateToken(user.ID, user.Email, user.Role, studentID, s.jwtSecret, s.jwtExpHours)
	if err != nil {
		return nil, helper.Internal(err, "Gagal menerbitkan token otentikasi")
	}

	return &LoginResponseData{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   exp,
		User: UserInfoDTO{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64, role string) (any, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil data pengguna")
	}
	if user == nil {
		return nil, helper.NotFound("Pengguna tidak ditemukan")
	}

	baseInfo := UserInfoDTO{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	if role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, helper.Internal(err, "Gagal mengambil profil mahasiswa")
		}
		if student == nil {
			return nil, helper.NotFound("Data profil mahasiswa tidak ditemukan")
		}
		return &MeMahasiswaResponse{
			UserInfoDTO: baseInfo,
			Student:     student,
		}, nil
	}

	return baseInfo, nil
}
