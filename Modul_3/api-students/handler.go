package main

import (
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

// -----------------------------------------------------------------------------
// StudentHandler — handler yang menerima repository melalui constructor injection
// -----------------------------------------------------------------------------

// StudentHandler menyimpan dependensi yang dibutuhkan oleh handler mahasiswa.
// Dengan menyimpan repo sebagai field, handler tidak bergantung pada variable global
// dan lebih mudah diuji secara terpisah.
type StudentHandler struct {
	repo repository.StudentRepository
}

// NewStudentHandler membuat StudentHandler baru dengan repository yang diberikan.
func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// -----------------------------------------------------------------------------
// Middleware
// -----------------------------------------------------------------------------

// requireJSON menolak request body dengan Content-Type selain application/json.
// Dipasang per-grup, bukan global, agar endpoint unggahan berkas tidak terdampak.
func requireJSON(c *fiber.Ctx) error {
	metodeBerbody := map[string]bool{
		fiber.MethodPost:  true,
		fiber.MethodPut:   true,
		fiber.MethodPatch: true,
	}
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}

// -----------------------------------------------------------------------------
// petaError: mengubah sentinel error dari repository → status HTTP
// -----------------------------------------------------------------------------

// petaError menerjemahkan error dari lapisan repository ke respons HTTP yang tepat.
// Memusatkan logika ini di satu tempat memudahkan perubahan di kemudian hari.
func (h *StudentHandler) petaError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	default:
		return fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server: "+err.Error())
	}
}

// -----------------------------------------------------------------------------
// Handler methods
// -----------------------------------------------------------------------------

// List menangani GET /api/v1/students.
func (h *StudentHandler) List(c *fiber.Ctx) error {
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
	// Batas atas limit = 100. Mencegah query terlalu besar yang bisa membebankan DB.
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

	list, total, err := h.repo.List(c.Context(), params)
	if err != nil {
		return h.petaError(c, err)
	}

	return ok(c, "daftar mahasiswa", list, &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: hitungTotalPages(total, limit),
	})
}

// Get menangani GET /api/v1/students/:id.
func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	s, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		return h.petaError(c, err)
	}
	return ok(c, "data mahasiswa", s, nil)
}

// Create menangani POST /api/v1/students.
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	errs := map[string]string{}
	nim := strings.TrimSpace(req.NIM)
	name := strings.TrimSpace(req.Name)

	if nim == "" {
		errs["nim"] = "nim wajib diisi"
	}
	if name == "" {
		errs["name"] = "name wajib diisi"
	}
	if req.IsActive == nil {
		errs["is_active"] = "is_active wajib diisi (true atau false)"
	}
	if gErr := validasiGrade(req.Grade); gErr != nil {
		errs["grade"] = gErr.Error()
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	isActive := *req.IsActive
	s, err := h.repo.Create(c.Context(), &model.Student{
		NIM:      nim,
		Name:     name,
		Grade:    req.Grade,
		IsActive: isActive,
	})
	if err != nil {
		return h.petaError(c, err)
	}

	return created(c, "/api/v1/students/"+floatToStr(float64(s.ID)), s)
}

// Replace menangani PUT /api/v1/students/:id.
func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req ReplaceRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	errs := map[string]string{}
	if req.NIM == nil || strings.TrimSpace(*req.NIM) == "" {
		errs["nim"] = "nim wajib diisi"
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		errs["name"] = "name wajib diisi"
	}
	if req.IsActive == nil {
		errs["is_active"] = "is_active wajib diisi"
	}
	if req.Grade == nil {
		errs["grade"] = "grade wajib diisi"
	} else if gErr := validasiGrade(*req.Grade); gErr != nil {
		errs["grade"] = gErr.Error()
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	s, err := h.repo.Replace(c.Context(), id, &model.Student{
		NIM:      strings.TrimSpace(*req.NIM),
		Name:     strings.TrimSpace(*req.Name),
		Grade:    *req.Grade,
		IsActive: *req.IsActive,
	})
	if err != nil {
		return h.petaError(c, err)
	}
	return ok(c, "data mahasiswa berhasil diperbarui", s, nil)
}

// Patch menangani PATCH /api/v1/students/:id.
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req PatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	errs := map[string]string{}
	fields := map[string]interface{}{}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			errs["nim"] = "nim tidak boleh kosong"
		} else {
			fields["nim"] = nim
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "name tidak boleh kosong"
		} else {
			fields["name"] = name
		}
	}
	if req.Grade != nil {
		if gErr := validasiGrade(*req.Grade); gErr != nil {
			errs["grade"] = gErr.Error()
		} else {
			fields["grade"] = *req.Grade
		}
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}

	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	s, err := h.repo.Patch(c.Context(), id, fields)
	if err != nil {
		return h.petaError(c, err)
	}
	return ok(c, "data mahasiswa berhasil diperbarui sebagian", s, nil)
}

// Delete menangani DELETE /api/v1/students/:id.
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.repo.Delete(c.Context(), id); err != nil {
		return h.petaError(c, err)
	}
	return noContent(c)
}
