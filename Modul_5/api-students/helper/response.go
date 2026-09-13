// Package helper menyediakan fungsi bantuan untuk menyusun respons HTTP
// yang seragam di seluruh aplikasi.
package helper

import (
	"github.com/gofiber/fiber/v2"
)

// WebResponse adalah amplop respons seragam untuk seluruh endpoint.
type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Meta menyimpan informasi paginasi untuk endpoint daftar.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success mengirim respons sukses dengan status dan amplop seragam.
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessList mengirim respons sukses dengan data daftar dan meta paginasi.
func SuccessList(c *fiber.Ctx, message string, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim 201 sekaligus memasang header Location.
func Created(c *fiber.Ctx, location string, data interface{}) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: "data berhasil ditambahkan",
		Data:    data,
	})
}

// NoContent mengirim 204 tanpa body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail mengirim respons gagal dengan amplop seragam.
func Fail(c *fiber.Ctx, status int, message string, errs ...interface{}) error {
	env := WebResponse{
		Success: false,
		Message: message,
	}
	if len(errs) > 0 && errs[0] != nil {
		env.Errors = errs[0]
	}
	return c.Status(status).JSON(env)
}
