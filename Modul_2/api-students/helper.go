package main

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ok mengirim respons sukses dengan amplop seragam.
func ok(c *fiber.Ctx, msg string, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{
		Success: true,
		Message: msg,
		Data:    data,
		Meta:    meta,
	})
}

// created mengirim 201 disertai header Location.
func created(c *fiber.Ctx, location string, data interface{}) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(Envelope{
		Success: true,
		Message: "data berhasil ditambahkan",
		Data:    data,
	})
}

// noContent mengirim 204 tanpa body.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// fail mengirim respons gagal dengan amplop seragam.
func fail(c *fiber.Ctx, status int, msg string, errs ...interface{}) error {
	env := Envelope{
		Success: false,
		Message: msg,
	}
	if len(errs) > 0 {
		env.Errors = errs[0]
	}
	return c.Status(status).JSON(env)
}

// parseID mengurai parameter :id dari URL dan mengembalikan error 400 jika bukan angka.
func parseID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	return id, nil
}

// floatToStr mengubah float64 menjadi string tanpa desimal untuk membentuk URL Location.
func floatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 0, 64)
}

// hitungTotalPages menghitung jumlah halaman berdasarkan total data dan limit.
func hitungTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}

// validasiGrade memastikan nilai grade berada dalam rentang 0–100.
func validasiGrade(grade float64) error {
	if grade < 0 || grade > 100 {
		return fmt.Errorf("grade harus antara 0 dan 100")
	}
	return nil
}
