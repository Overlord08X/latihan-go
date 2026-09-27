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

// WebResponse adalah amplop respons seragam untuk endpoint sukses.
type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// Success mengirim respons sukses dengan status dan amplop seragam.
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessCursor mengirim respons sukses dengan data daftar dan CursorMeta (Langkah 7).
func SuccessCursor(c *fiber.Ctx, message string, data interface{}, meta *model.CursorMeta) error {
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

// RequestID mengambil request ID dari Locals Fiber.
func RequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestid").(string); ok && id != "" {
		return id
	}
	return ""
}
