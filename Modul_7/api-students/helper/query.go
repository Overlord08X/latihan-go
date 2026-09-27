// Package helper menyediakan fungsi bantuan untuk membaca query string
// dan parameter URL.
package helper

import (
	"strconv"
	"strings"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

// ParseCursorQuery membaca parameter query cursor pagination untuk users:
// - limit (default 10, max 100, jika tidak valid -> 400 BAD_REQUEST)
// - cursor (base64 string -> DecodeCursor, jika rusak -> 400 BAD_REQUEST)
// - is_active (harus 'true' atau 'false', jika nilai lain -> 400 BAD_REQUEST)
// - search (string pencarian username)
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	q := model.CursorQuery{
		Limit: DefaultLimit,
	}

	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			return q, BadRequest("limit harus berupa angka positif")
		}
		if l > MaxLimit {
			l = MaxLimit // Pembatasan maksimal limit
		}
		q.Limit = l
	}

	if cursorStr := strings.TrimSpace(c.Query("cursor")); cursorStr != "" {
		cur, err := DecodeCursor(cursorStr)
		if err != nil {
			return q, BadRequest("cursor tidak valid")
		}
		q.After = &cur
	}

	if activeStr := strings.TrimSpace(c.Query("is_active")); activeStr != "" {
		lower := strings.ToLower(activeStr)
		if lower != "true" && lower != "false" {
			return q, BadRequest("is_active harus bernilai true atau false")
		}
		isActive := lower == "true"
		q.IsActive = &isActive
	}

	q.Search = strings.TrimSpace(c.Query("search"))

	return q, nil
}

// ParseStudentCursorQuery membaca parameter query cursor pagination untuk students:
// - limit (default 10, max 100)
// - cursor (DecodeCursor)
// - is_active (true / false)
// - search (string pencarian nama/nim)
func ParseStudentCursorQuery(c *fiber.Ctx) (model.StudentCursorQuery, error) {
	q := model.StudentCursorQuery{
		Limit: DefaultLimit,
	}

	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			return q, BadRequest("limit harus berupa angka positif")
		}
		if l > MaxLimit {
			l = MaxLimit
		}
		q.Limit = l
	}

	if cursorStr := strings.TrimSpace(c.Query("cursor")); cursorStr != "" {
		cur, err := DecodeCursor(cursorStr)
		if err != nil {
			return q, BadRequest("cursor tidak valid")
		}
		q.After = &cur
	}

	if activeStr := strings.TrimSpace(c.Query("is_active")); activeStr != "" {
		lower := strings.ToLower(activeStr)
		if lower != "true" && lower != "false" {
			return q, BadRequest("is_active harus bernilai true atau false")
		}
		isActive := lower == "true"
		q.IsActive = &isActive
	}

	q.Search = strings.TrimSpace(c.Query("search"))

	return q, nil
}
