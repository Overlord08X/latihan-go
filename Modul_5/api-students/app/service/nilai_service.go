package service

import (
	"errors"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type NilaiService struct {
	repo repository.NilaiRepository
}

func NewNilaiService(repo repository.NilaiRepository) *NilaiService {
	return &NilaiService{repo: repo}
}

// Create menangani POST /nilais
func (s *NilaiService) Create(c *fiber.Ctx) error {
	var req model.CreateNilaiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}

	// Pisahkan aturan bisnis murni ke NilaiRules
	validationErrs, err := ValidateCreateNilai(&req)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", validationErrs)
	}

	// Mapping dari Request ke Entity
	nilai := &model.Nilai{
		IDStudent:  *req.IDStudent,
		NamaMatkul: req.NamaMatkul,
		Nilai:      *req.Nilai,
	}

	createdNilai, err := s.repo.Create(c.Context(), nilai)
	if err != nil {
		return s.petaError(c, err)
	}

	return helper.Created(c, "", createdNilai)
}

// GetByStudentNIM menangani GET /students/:nim/nilais
func (s *NilaiService) GetByStudentNIM(c *fiber.Ctx) error {
	nim := c.Params("nim")

	nilais, err := s.repo.GetByStudentNIM(c.Context(), nim)
	if err != nil {
		return s.petaError(c, err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar nilai mahasiswa", nilais)
}

func (s *NilaiService) petaError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}
}
