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
		return helper.BadRequest("format JSON tidak valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	nilai := &model.Nilai{
		IDStudent:  *req.IDStudent,
		NamaMatkul: req.NamaMatkul,
		Nilai:      *req.Nilai,
	}

	createdNilai, err := s.repo.Create(c.Context(), nilai)
	if err != nil {
		return s.petaError(err)
	}

	return helper.Created(c, "", createdNilai)
}

// GetByStudentNIM menangani GET /students/:nim/nilais
func (s *NilaiService) GetByStudentNIM(c *fiber.Ctx) error {
	nim := c.Params("nim")

	nilais, err := s.repo.GetByStudentNIM(c.Context(), nim)
	if err != nil {
		return s.petaError(err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar nilai mahasiswa", nilais)
}

func (s *NilaiService) petaError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")
	default:
		return helper.Internal(err)
	}
}
