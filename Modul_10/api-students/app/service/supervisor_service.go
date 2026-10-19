package service

import (
	"errors"
	"strconv"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type SupervisorService struct {
	repo repository.SupervisorRepository
	tx   *repository.TxManager
}

func NewSupervisorService(repo repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{repo: repo, tx: tx}
}

// ---------- GET /supervisors ----------
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar dosen pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "dosen pembimbing ditemukan", sp)
}

// ---------- POST /supervisors/:id/assign ----------
// Menetapkan mahasiswa ke dosen pembimbing dengan UPDATE bersyarat dan pessimistic lock pada baris mahasiswa.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id dosen pembimbing harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var updatedSupervisor model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		txRepo := s.repo.WithTx(tx)

		// 1. Kunci baris mahasiswa di PostgreSQL untuk mencegah penetapan ganda secara bersamaan.
		if err := txRepo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Naikkan jumlah bimbingan secara atomik (UPDATE bersyarat kuota).
		var err error
		updatedSupervisor, err = txRepo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		// 3. Catat bimbingan (PRIMARY KEY student_id menjamin 1 mahasiswa = 1 pembimbing).
		return txRepo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.CreatedWithMsg(c, "berhasil menetapkan dosen pembimbing", updatedSupervisor,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
// Mengganti pembimbing secara atomik dan bebas deadlock (kunci berurutan).
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var newSupervisor model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		txRepo := s.repo.WithTx(tx)

		// 1. Kunci mahasiswa
		if err := txRepo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Ambil pembimbing saat ini
		currentSupervisorID, err := txRepo.GetSupervisionByStudent(ctx, req.StudentID)
		if err != nil {
			return err
		}

		if currentSupervisorID == req.ToSupervisorID {
			return helper.NewError(fiber.StatusBadRequest, "SAME_SUPERVISOR",
				"dosen pembimbing tujuan tidak boleh sama dengan dosen saat ini")
		}

		// 3. Kunci kedua supervisor dalam urutan ID menaik untuk mencegah deadlock
		if err := txRepo.LockSupervisors(ctx, currentSupervisorID, req.ToSupervisorID); err != nil {
			return err
		}

		// 4. Hapus bimbingan lama & kurangi kuota pembimbing lama
		if err := txRepo.DeleteSupervision(ctx, req.StudentID); err != nil {
			return err
		}
		if _, err := txRepo.DecrementSupervised(ctx, currentSupervisorID); err != nil {
			return err
		}

		// 5. Tambah kuota pembimbing baru secara atomik (bersyarat)
		newSupervisor, err = txRepo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}

		// 6. Catat bimbingan baru
		return txRepo.InsertSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti dosen pembimbing", newSupervisor)
}

// ---------- PATCH /supervisors/:id ----------
// Ubah kuota dosen pembimbing dengan optimistic locking.
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.repo.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota dosen pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	var appErr *helper.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, repository.ErrSupervisorNotFound):
		return helper.NotFound("dosen pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.NotFound("data mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota bimbingan dosen sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki dosen pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki dosen pembimbing")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa yang sedang dibimbing")
	default:
		return helper.Internal(err)
	}
}
