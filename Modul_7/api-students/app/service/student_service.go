package service

import (
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// StudentService menyimpan dependensi yang dibutuhkan oleh handler mahasiswa.
type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

// NewStudentService membuat StudentService baru dengan repository dan permission set.
func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// translateStudentError menerjemahkan error dari lapisan repository ke AppError.
func translateStudentError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah digunakan")
	default:
		return helper.Internal(err)
	}
}

// List menangani GET /api/v1/students dengan pagination berbasis cursor dan content negotiation (Tugas Mandiri D.3 & D.4).
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// 1. Content Negotiation
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	// 2. Parse Cursor Query parameter
	q, err := helper.ParseStudentCursorQuery(c)
	if err != nil {
		return err
	}

	// 3. Keyset Pagination
	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	// Layani format CSV
	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	// Potong limit+1 baris untuk menentukan has_more
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar mahasiswa berhasil diambil", rows, meta)
}

// Get menangani GET /api/v1/students/:id.
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	var ownerID int
	if student.OwnerID != nil {
		ownerID = *student.OwnerID
	}

	// Pemeriksaan kepemilikan data:
	// Pemilik data (owner_id == user.UserID) selalu boleh; selain itu butuh student:read:any.
	if !CanAccessStudent(user, ownerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data mahasiswa lain")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa", student)
}

// Create menangani POST /api/v1/students.
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	student := model.Student{
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: *req.IsActive,
		OwnerID:  &user.UserID, // server-assigned owner_id
	}

	created, err := s.repo.Create(ctx, &student)
	if err != nil {
		return translateStudentError(err)
	}

	location := "/api/v1/students/" + strconv.Itoa(created.ID)
	return helper.Created(c, location, created)
}

// Replace menangani PUT /api/v1/students/:id.
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	var ownerID int
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}

	if !CanAccessStudent(user, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa lain")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	student := model.Student{
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    *req.Grade,
		IsActive: *req.IsActive,
	}

	updated, err := s.repo.Replace(ctx, id, &student)
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui", updated)
}

// Patch menangani PATCH /api/v1/students/:id (Tugas Mandiri D.2).
func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}

	var ownerID int
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}

	if !CanAccessStudent(user, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data mahasiswa lain")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	// 1. Validasi deklaratif tag struct (memeriksa aturan omitnil, min, nim, dll)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// 2. Validasi hubungan antar field: body tidak boleh kosong sama sekali
	if IsEmptyPatchStudent(req) {
		return helper.BadRequest("setidaknya satu field harus dikirim")
	}

	patched := ApplyPatchStudent(*existing, req)

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

	updated, err := s.repo.Patch(ctx, id, fields)
	if err != nil {
		return translateStudentError(err)
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui sebagian", updated)
}

// Delete menangani DELETE /api/v1/students/:id.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateStudentError(err)
	}

	return helper.NoContent(c)
}
