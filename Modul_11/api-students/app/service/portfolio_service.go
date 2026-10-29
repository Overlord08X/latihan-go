package service

import (
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// ---------- GET /students/:id/portfolio ----------
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// 1. Verifikasi mahasiswa ada di PostgreSQL
	if _, err := s.students.GetByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	// 2. Ambil portofolio dari MongoDB
	portfolio, err := s.portfolios.GetByStudentID(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio mahasiswa berhasil diambil", portfolio)
}

// ---------- PUT /students/:id/portfolio ----------
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// 1. Verifikasi mahasiswa ada di PostgreSQL
	if _, err := s.students.GetByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	var req model.PutPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		req.Projects[i].Title = strings.TrimSpace(req.Projects[i].Title)
		req.Projects[i].Technologies = NormalizeTags(req.Projects[i].Technologies)
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.Upsert(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

// ---------- POST /students/:id/portfolio/projects ----------
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// 1. Verifikasi mahasiswa ada di PostgreSQL
	if _, err := s.students.GetByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Technologies = NormalizeTags(req.Technologies)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.portfolios.AddProject(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan ke portofolio", updated)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// 1. Verifikasi mahasiswa ada di PostgreSQL
	if _, err := s.students.GetByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("projectId harus diisi")
	}

	updated, err := s.portfolios.DeleteProject(ctx, studentID, projectID)
	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus dari portofolio", updated)
}

// ---------- GET /portfolios/stats/technologies ----------
func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TechnologyStats(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "statistik 10 teknologi terpopuler berhasil diambil", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id tidak valid")
	case errors.Is(err, repository.ErrPortfolioNotFound):
		return helper.NotFound("portofolio belum dibuat untuk mahasiswa ini")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan di dalam portofolio")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT_EXCEEDED",
			"jumlah proyek sudah mencapai batas maksimal 20")
	default:
		return helper.Internal(err)
	}
}
