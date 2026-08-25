package main

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// -----------------------------------------------------------------------------
// In-memory store
// -----------------------------------------------------------------------------

// store adalah "database" in-memory yang dilindungi oleh mutex untuk keamanan
// akses bersamaan (concurrent access).
var (
	mu       sync.RWMutex
	students = make(map[int]*Student)
	nextID   = 1
)

// -----------------------------------------------------------------------------
// Middleware
// -----------------------------------------------------------------------------

// requireJSON menolak request yang berisi body tetapi Content-Type-nya bukan
// application/json. Status 415 (Unsupported Media Type) lebih tepat dari 400
// karena masalahnya ada pada format media, bukan isi body.
//
// Middleware ini hanya dipasang pada grup /students karena endpoint yang
// menerima unggahan berkas di masa depan (multipart/form-data) tidak boleh
// ikut tertolak.
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
// Handlers
// -----------------------------------------------------------------------------

// listStudents menangani GET /api/v1/students.
// Mendukung: paginasi, pencarian nama, pengurutan, dan filter is_active.
func listStudents(c *fiber.Ctx) error {
	// --- query string dengan nilai bawaan yang aman ---
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := strings.ToLower(c.Query("search"))
	sortBy := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))
	isActiveQ := c.Query("is_active")

	// Validasi & sanitasi input
	if page < 1 {
		page = 1
	}
	// Batas atas limit = 100. Alasan: melindungi server dari payload respons
	// yang terlalu besar dan serangan semacam ?limit=999999.
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Daftar putih field yang boleh diurutkan, mencegah SQL-injection-like
	// pada nama field (meski di sini in-memory).
	allowedSort := map[string]bool{"id": true, "nim": true, "name": true, "grade": true}
	if !allowedSort[sortBy] {
		sortBy = "id"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	mu.RLock()
	list := make([]*Student, 0, len(students))
	for _, s := range students {
		list = append(list, s)
	}
	mu.RUnlock()

	// --- filter ---
	filtered := make([]*Student, 0, len(list))
	for _, s := range list {
		// Filter is_active
		if isActiveQ == "true" && !s.IsActive {
			continue
		}
		if isActiveQ == "false" && s.IsActive {
			continue
		}
		// Filter pencarian nama (case-insensitive)
		if search != "" && !strings.Contains(strings.ToLower(s.Name), search) {
			continue
		}
		filtered = append(filtered, s)
	}

	// --- sort ---
	sort.Slice(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		var less bool
		switch sortBy {
		case "nim":
			less = a.NIM < b.NIM
		case "name":
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case "grade":
			less = a.Grade < b.Grade
		default:
			less = a.ID < b.ID
		}
		if order == "desc" {
			return !less
		}
		return less
	})

	// --- paginasi ---
	total := len(filtered)
	totalPages := hitungTotalPages(total, limit)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	paged := filtered[start:end]

	return ok(c, "daftar mahasiswa", paged, &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// getStudent menangani GET /api/v1/students/:id.
func getStudent(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	mu.RLock()
	s, ada := students[id]
	mu.RUnlock()

	if !ada {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	return ok(c, "data mahasiswa", s, nil)
}

// createStudent menangani POST /api/v1/students.
func createStudent(c *fiber.Ctx) error {
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	// Validasi field wajib
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
	if err := validasiGrade(req.Grade); err != nil {
		errs["grade"] = err.Error()
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	mu.Lock()
	// Cek NIM ganda (409 Conflict)
	for _, s := range students {
		if strings.EqualFold(s.NIM, nim) {
			mu.Unlock()
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
	}
	id := nextID
	nextID++
	isActive := *req.IsActive
	s := &Student{
		ID:       id,
		NIM:      nim,
		Name:     name,
		Grade:    req.Grade,
		IsActive: isActive,
	}
	students[id] = s
	mu.Unlock()

	return created(c, "/api/v1/students/"+floatToStr(float64(id)), s)
}

// replaceStudent menangani PUT /api/v1/students/:id.
// PUT bersifat idempoten: request yang sama dijalankan berkali-kali menghasilkan
// keadaan yang sama. Seluruh field wajib hadir.
func replaceStudent(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req ReplaceRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	// PUT: semua field wajib hadir
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
	} else if err := validasiGrade(*req.Grade); err != nil {
		errs["grade"] = err.Error()
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	nim := strings.TrimSpace(*req.NIM)
	name := strings.TrimSpace(*req.Name)

	mu.Lock()
	defer mu.Unlock()

	if _, ada := students[id]; !ada {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	// Cek NIM ganda dengan mahasiswa lain
	for sid, s := range students {
		if sid != id && strings.EqualFold(s.NIM, nim) {
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan oleh mahasiswa lain")
		}
	}

	students[id] = &Student{
		ID:       id,
		NIM:      nim,
		Name:     name,
		Grade:    *req.Grade,
		IsActive: *req.IsActive,
	}
	return ok(c, "data mahasiswa berhasil diperbarui", students[id], nil)
}

// patchStudent menangani PATCH /api/v1/students/:id.
// PATCH hanya mengubah field yang dikirim; field yang tidak ada tetap sama.
// Berbeda dari PUT yang mengganti seluruh resource.
func patchStudent(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req PatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body bukan JSON yang sah")
	}

	mu.Lock()
	defer mu.Unlock()

	s, ada := students[id]
	if !ada {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	// Validasi field yang dikirim saja
	errs := map[string]string{}
	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			errs["nim"] = "nim tidak boleh kosong"
		} else {
			// Cek NIM ganda
			for sid, other := range students {
				if sid != id && strings.EqualFold(other.NIM, nim) {
					return fail(c, fiber.StatusConflict, "NIM sudah digunakan oleh mahasiswa lain")
				}
			}
			s.NIM = nim
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "name tidak boleh kosong"
		} else {
			s.Name = name
		}
	}
	if req.Grade != nil {
		if err := validasiGrade(*req.Grade); err != nil {
			errs["grade"] = err.Error()
		} else {
			s.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		s.IsActive = *req.IsActive
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", s, nil)
}

// deleteStudent menangani DELETE /api/v1/students/:id.
// Mengembalikan 204 tanpa body.
func deleteStudent(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ada := students[id]; !ada {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	delete(students, id)
	return noContent(c)
}

// initSampleData mengisi beberapa data awal agar endpoint GET bisa langsung diuji.
func initSampleData() {
	aktif := true
	nonaktif := false
	_ = time.Now() // pastikan paket time dipakai

	samples := []struct {
		nim      string
		name     string
		grade    float64
		isActive *bool
	}{
		{"2023001", "Andi Saputra", 85.5, &aktif},
		{"2023002", "Budi Santoso", 72.0, &aktif},
		{"2023003", "Citra Dewi", 91.0, &nonaktif},
	}

	for _, d := range samples {
		isActive := *d.isActive
		s := &Student{
			ID:       nextID,
			NIM:      d.nim,
			Name:     d.name,
			Grade:    d.grade,
			IsActive: isActive,
		}
		students[nextID] = s
		nextID++
	}
}
