// Package helper menyediakan fungsi bantuan untuk menyusun respons HTTP
// yang seragam di seluruh aplikasi.
package helper

import (
	"context"
	"strconv"
	"time"

	"api-students/app/model"

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

// FailValidation mengirim respons 422 Unprocessable Entity dengan rincian error per field.
func FailValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errors,
	})
}

// ParamID mengambil param :id dan memvalidasi bahwa nilainya adalah angka positif (> 0).
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// CurrentUser mengambil data pengguna yang terautentikasi dari Locals Fiber.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	rawID := c.Locals("userID")
	rawRole := c.Locals("role")
	rawUsername := c.Locals("username")

	if rawID == nil || rawRole == nil {
		return model.AuthUser{}, false
	}

	var userID int
	switch v := rawID.(type) {
	case int:
		userID = v
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return model.AuthUser{}, false
		}
		userID = id
	default:
		return model.AuthUser{}, false
	}

	role, ok := rawRole.(string)
	if !ok || role == "" {
		return model.AuthUser{}, false
	}

	username, _ := rawUsername.(string)

	return model.AuthUser{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, true
}

// RequestContext membuat context dengan timeout 10 detik dari request Fiber.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 10*time.Second)
}
