// Package service berisi handler HTTP yang membaca request, memanggil
// business rules, dan menyusun respons. Handler tidak mengandung logika
// bisnis maupun query SQL secara langsung.
package service

import (
	"errors"
	"strconv"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// StudentService menyimpan dependensi yang dibutuhkan oleh handler mahasiswa.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService membuat StudentService baru dengan repository yang diberikan.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// petaError menerjemahkan error dari lapisan repository ke respons HTTP.
func (s *StudentService) petaError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server: "+err.Error())
	}
}

// List menangani GET /api/v1/students.
func (s *StudentService) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search")
	sortBy := c.Query("sort", "id")
	order := c.Query("order", "asc")
	isActiveQ := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var isActive *bool
	if isActiveQ == "true" {
		v := true
		isActive = &v
	} else if isActiveQ == "false" {
		v := false
		isActive = &v
	}

	params := repository.ListParams{
		Page: page, Limit: limit,
		Search: search, SortBy: sortBy, Order: order,
		IsActive: isActive,
	}

	list, total, err := s.repo.List(c.Context(), params)
	if err != nil {
		return s.petaError(c, err)
	}

	return helper.SuccessList(c, "daftar mahasiswa", list, &helper.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: helper.CountTotalPages(total, limit),
	})
}

// Get menangani GET /api/v1/students/:id.
func (s *StudentService) Get(c *fiber.Ctx) error {
	id, err := helper.ParseID(c)
	if err != nil {
		return err
	}
	student, err := s.repo.GetByID(c.Context(), id)
	if err != nil {
		return s.petaError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "data mahasiswa", student)
}

// Create menangani POST /api/v1/students.
func (s *StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	student, errs := ValidateCreate(req)
	if len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	created, err := s.repo.Create(c.Context(), &student)
	if err != nil {
		return s.petaError(c, err)
	}

	location := "/api/v1/students/" + strconv.Itoa(created.ID)
	return helper.Created(c, location, created)
}

// Replace menangani PUT /api/v1/students/:id.
func (s *StudentService) Replace(c *fiber.Ctx) error {
	id, err := helper.ParseID(c)
	if err != nil {
		return err
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	student, errs := ValidateReplace(req)
	if len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	updated, err := s.repo.Replace(c.Context(), id, &student)
	if err != nil {
		return s.petaError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui", updated)
}

// Patch menangani PATCH /api/v1/students/:id.
func (s *StudentService) Patch(c *fiber.Ctx) error {
	id, err := helper.ParseID(c)
	if err != nil {
		return err
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	// Ambil data saat ini dari DB
	current, err := s.repo.GetByID(c.Context(), id)
	if err != nil {
		return s.petaError(c, err)
	}

	// Terapkan perubahan melalui business rules
	patched, errs := ApplyPatch(*current, req)
	if len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	// Simpan perubahan ke DB
	fields := map[string]interface{}{}
	if req.NIM != nil {
		fields["nim"] = patched.NIM
	}
	if req.Name != nil {
		fields["name"] = patched.Name
	}
	if req.Grade != nil {
		fields["grade"] = patched.Grade
	}
	if req.IsActive != nil {
		fields["is_active"] = patched.IsActive
	}

	updated, err := s.repo.Patch(c.Context(), id, fields)
	if err != nil {
		return s.petaError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui sebagian", updated)
}

// Delete menangani DELETE /api/v1/students/:id.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	id, err := helper.ParseID(c)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(c.Context(), id); err != nil {
		return s.petaError(c, err)
	}
	return helper.NoContent(c)
}
