// Package helper menyediakan fungsi bantuan untuk membaca query string
// dan parameter URL.
package helper

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ParseID mengurai parameter :id dari URL.
// Mengembalikan error 400 bila bukan angka positif.
func ParseID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	return id, nil
}

// CountTotalPages menghitung jumlah halaman berdasarkan total data dan limit.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}

// ValidateGrade memastikan nilai grade berada dalam rentang 0–100.
func ValidateGrade(grade float64) error {
	if grade < 0 || grade > 100 {
		return fmt.Errorf("grade harus antara 0 dan 100")
	}
	return nil
}

// FloatToStr mengubah float64 menjadi string integer untuk membentuk URL Location.
func FloatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 0, 64)
}
