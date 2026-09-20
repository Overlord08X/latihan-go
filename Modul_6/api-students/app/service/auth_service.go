package service

import (
	"errors"
	"os"
	"strconv"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/jwtutil"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost adalah cost yang digunakan untuk hashing password.
// Cost 12 dipilih: cukup lambat untuk menahan brute force (~300ms per hash)
// tetapi tidak terlalu lambat bagi UX pada server modern.
const bcryptCost = 12

// AuthService menangani seluruh alur autentikasi.
type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	perms     *helper.PermissionSet
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, perms *helper.PermissionSet) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, perms: perms}
}

// Register menangani POST /api/v1/auth/register
func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}

	// Validasi via business rules murni (tanpa Fiber)
	validationErrs, err := ValidateRegister(req.Username, req.Email, req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", validationErrs)
	}

	// Hash password dengan bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	// Role SELALU ditentukan server — mass assignment prevention
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	created, err := s.userRepo.Create(c.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	// Kembalikan UserProfile — TANPA password (field password tidak ada di struct ini)
	profile := &model.UserProfile{
		ID:        created.ID,
		Username:  created.Username,
		Email:     created.Email,
		Role:      created.Role,
		CreatedAt: created.CreatedAt,
	}

	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", profile)
}

// Login menangani POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}

	// Ambil user dari database
	user, err := s.userRepo.GetByUsername(c.Context(), req.Username)

	// Selalu jalankan bcrypt comparison, bahkan jika user tidak ada.
	// Ini mencegah user enumeration via timing attack:
	// waktu respons "username tidak ada" dan "password salah" menjadi serupa.
	dummyHash := "$2a$12$qkCtBjxYl7HKDCUeFzbnLOXvr9Ac9QSqGijezb6mCqzOOjb.OzRpu" // hash dari "dummy"
	compareTarget := dummyHash
	if err == nil {
		compareTarget = user.Password
	}

	// Selalu lakukan perbandingan hash — ini yang membuat waktunya serupa
	compareErr := bcrypt.CompareHashAndPassword([]byte(compareTarget), []byte(req.Password))

	// Pesan error SAMA PERSIS untuk "user tidak ditemukan" dan "password salah"
	// Ini menutup kerentanan user enumeration
	if err != nil || compareErr != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password tidak valid")
	}

	// Baca TTL dari environment
	accessTTL := getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshTTLDays := getEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	// Buat access token (JWT)
	accessToken, err := jwtutil.GenerateAccessToken(user.ID, user.Username, user.Role, accessTTL)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	// Buat refresh token (random string)
	refreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	// Simpan hash refresh token ke database
	expiresAt := time.Now().Add(time.Duration(refreshTTLDays) * 24 * time.Hour)
	tokenHash := jwtutil.HashToken(refreshToken)
	if err := s.tokenRepo.Save(c.Context(), user.ID, tokenHash, expiresAt); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    accessTTL * 60, // dalam detik
	})
}

// Refresh menangani POST /api/v1/auth/refresh
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh_token wajib diisi")
	}

	tokenHash := jwtutil.HashToken(req.RefreshToken)

	// FindAndRevoke: cari token, validasi, lalu revoke (rotasi)
	userID, err := s.tokenRepo.FindAndRevoke(c.Context(), tokenHash)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah kedaluwarsa")
	}

	// Ambil data user untuk membuat token baru
	user, err := s.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "pengguna tidak ditemukan")
	}

	accessTTL := getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshTTLDays := getEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	// Buat pasangan token baru
	newAccessToken, err := jwtutil.GenerateAccessToken(user.ID, user.Username, user.Role, accessTTL)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	newRefreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	expiresAt := time.Now().Add(time.Duration(refreshTTLDays) * 24 * time.Hour)
	newTokenHash := jwtutil.HashToken(newRefreshToken)
	if err := s.tokenRepo.Save(c.Context(), user.ID, newTokenHash, expiresAt); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}

	return helper.Success(c, fiber.StatusOK, "token diperbarui", fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    accessTTL * 60,
	})
}

// Logout menangani POST /api/v1/auth/logout
func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh_token wajib diisi")
	}

	tokenHash := jwtutil.HashToken(req.RefreshToken)
	// Abaikan error — bila token sudah dicabut atau tidak ditemukan, tetap 200
	_, _ = s.tokenRepo.FindAndRevoke(c.Context(), tokenHash)

	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

// Me menangani GET /api/v1/auth/me
// Middleware RequireAuth sudah berjalan sebelum handler ini, sehingga token pasti valid.
func (s *AuthService) Me(c *fiber.Ctx) error {
	userAuth, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "identitas tidak ditemukan")
	}

	user, err := s.userRepo.GetByID(c.Context(), userAuth.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "pengguna tidak ditemukan")
	}

	profile := &model.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	var permissions []string
	if s.perms != nil {
		permissions = s.perms.PermissionsOf(user.Role)
	} else {
		permissions = []string{}
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"user":        profile,
		"permissions": permissions,
	})
}

// getEnvInt membaca environment variable sebagai integer dengan nilai default.
func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}
