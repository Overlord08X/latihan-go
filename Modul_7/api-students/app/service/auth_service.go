package service

import (
	"errors"
	"os"
	"strconv"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
	"api-students/jwtutil"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost adalah cost yang digunakan untuk hashing password.
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

// Register menangani POST /api/v1/auth/register (Langkah 6).
func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	// Validasi deklaratif menggunakan tag struct (Langkah 6)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Hash password dengan bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return helper.Internal(err)
	}

	// Role SELALU ditentukan server — mass assignment prevention
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
		IsActive: true,
	}

	created, err := s.userRepo.Create(c.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("username atau email sudah digunakan")
		}
		return helper.Internal(err)
	}

	profile := &model.UserProfile{
		ID:        created.ID,
		Username:  created.Username,
		Email:     created.Email,
		Role:      created.Role,
		IsActive:  created.IsActive,
		CreatedAt: created.CreatedAt,
	}

	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", profile)
}

// Login menangani POST /api/v1/auth/login.
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Ambil user dari database
	user, err := s.userRepo.GetByUsername(c.Context(), req.Username)

	// Selalu jalankan bcrypt comparison untuk timing safety (mencegah user enumeration)
	dummyHash := "$2a$12$qkCtBjxYl7HKDCUeFzbnLOXvr9Ac9QSqGijezb6mCqzOOjb.OzRpu"
	compareTarget := dummyHash
	if err == nil {
		compareTarget = user.Password
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(compareTarget), []byte(req.Password))

	if err != nil || compareErr != nil {
		return helper.Unauthorized("username atau password tidak valid")
	}

	accessTTL := getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshTTLDays := getEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	accessToken, err := jwtutil.GenerateAccessToken(user.ID, user.Username, user.Role, accessTTL)
	if err != nil {
		return helper.Internal(err)
	}

	refreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return helper.Internal(err)
	}

	expiresAt := time.Now().Add(time.Duration(refreshTTLDays) * 24 * time.Hour)
	tokenHash := jwtutil.HashToken(refreshToken)
	if err := s.tokenRepo.Save(c.Context(), user.ID, tokenHash, expiresAt); err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    accessTTL * 60,
	})
}

// Refresh menangani POST /api/v1/auth/refresh.
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.BadRequest("refresh_token wajib diisi")
	}

	tokenHash := jwtutil.HashToken(req.RefreshToken)
	userID, err := s.tokenRepo.FindAndRevoke(c.Context(), tokenHash)
	if err != nil {
		return helper.Unauthorized("refresh token tidak valid atau sudah kedaluwarsa")
	}

	user, err := s.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return helper.Unauthorized("pengguna tidak ditemukan")
	}

	accessTTL := getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshTTLDays := getEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	newAccessToken, err := jwtutil.GenerateAccessToken(user.ID, user.Username, user.Role, accessTTL)
	if err != nil {
		return helper.Internal(err)
	}

	newRefreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return helper.Internal(err)
	}

	expiresAt := time.Now().Add(time.Duration(refreshTTLDays) * 24 * time.Hour)
	newTokenHash := jwtutil.HashToken(newRefreshToken)
	if err := s.tokenRepo.Save(c.Context(), user.ID, newTokenHash, expiresAt); err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "token diperbarui", fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    accessTTL * 60,
	})
}

// Logout menangani POST /api/v1/auth/logout.
func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.BadRequest("refresh_token wajib diisi")
	}

	tokenHash := jwtutil.HashToken(req.RefreshToken)
	_, _ = s.tokenRepo.FindAndRevoke(c.Context(), tokenHash)

	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

// Me menangani GET /api/v1/auth/me.
func (s *AuthService) Me(c *fiber.Ctx) error {
	userAuth, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("identitas tidak ditemukan")
	}

	user, err := s.userRepo.GetByID(c.Context(), userAuth.UserID)
	if err != nil {
		return helper.NotFound("pengguna tidak ditemukan")
	}

	profile := &model.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
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
