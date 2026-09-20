package service

import (
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{repo: repo, perms: perms}
}

// translateError menerjemahkan repository error ke Fiber error response
func (s *UserService) translateError(c *fiber.Ctx, err error, defaultMsg string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "username atau email sudah terdaftar")
	}
	return helper.Fail(c, fiber.StatusInternalServerError, defaultMsg)
}

// List mengambil daftar seluruh user (GET /users)
func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	users, err := s.repo.List(ctx)
	if err != nil {
		return s.translateError(c, err, "gagal mengambil daftar user")
	}

	// Jangan kembalikan password
	var profiles []model.UserProfile
	for _, u := range users {
		profiles = append(profiles, model.UserProfile{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}

	return helper.Success(c, fiber.StatusOK, "daftar user berhasil diambil", profiles)
}

// Get mengambil detail user berdasarkan ID (GET /users/:id)
func (s *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Pemeriksaan hak akses dilakukan SEBELUM data diambil (mencegah timing attack).
	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return s.translateError(c, err, "gagal mengambil data user")
	}

	profile := model.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", profile)
}

// AssignRole mengubah role user (PATCH /users/:id/role)
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return s.translateError(c, err, "gagal mengubah role user")
	}

	profile := model.UserProfile{
		ID:        result.ID,
		Username:  result.Username,
		Email:     result.Email,
		Role:      result.Role,
		CreatedAt: result.CreatedAt,
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", profile)
}

// Delete menghapus user (DELETE /users/:id)
func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Punya permission menghapus tidak berarti boleh menghapus dirinya sendiri.
	if current.UserID == id {
		return helper.Fail(c, fiber.StatusForbidden, "tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return s.translateError(c, err, "gagal menghapus user")
	}

	return helper.NoContent(c)
}
