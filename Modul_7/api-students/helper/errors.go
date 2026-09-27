package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Kode error yang stabil dan dapat dibaca mesin.
//
// Message boleh berubah kapan saja — ia ditulis untuk manusia.
// Code TIDAK boleh berubah — ia bagian dari kontrak API, karena client
// menuliskan percabangan berdasarkan nilainya.
const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeNotAcceptable    = "NOT_ACCEPTABLE"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeInternal         = "INTERNAL_ERROR"
)

// AppError adalah satu-satunya bentuk kegagalan yang dikenal aplikasi ini.
//
// Perhatikan bahwa ia TIDAK menyentuh fiber.Ctx. Sebuah error hanya
// menggambarkan apa yang salah; urusan menuliskannya sebagai response
// diserahkan sepenuhnya kepada ErrorHandler terpusat.
type AppError struct {
	Status  int               // status HTTP yang akan dikirim
	Code    string            // kode stabil untuk client
	Message string            // penjelasan untuk manusia
	Fields  map[string]string // detail per-field, khusus kegagalan validasi
	cause   error             // error asli, untuk log — tidak pernah dikirim ke client
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap membuat errors.Is dan errors.As tetap dapat menembus AppError
// untuk menemukan error asli di bawahnya.
func (e *AppError) Unwrap() error { return e.cause }

// Cause mengembalikan error asli (untuk keperluan logging terpusat).
func (e *AppError) Cause() error { return e.cause }

func BadRequest(message string) *AppError {
	return &AppError{Status: fiber.StatusBadRequest, Code: CodeBadRequest, Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Status: fiber.StatusUnauthorized, Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: fiber.StatusForbidden, Code: CodeForbidden, Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Status: fiber.StatusNotFound, Code: CodeNotFound, Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Status: fiber.StatusConflict, Code: CodeConflict, Message: message}
}

// Validation mengembalikan AppError untuk kegagalan validasi.
// Perbaikan Bug Perilaku #1: Status wajib 422 (fiber.StatusUnprocessableEntity), bukan 400.
func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Code:    CodeValidation,
		Message: "validasi gagal",
		Fields:  fields,
	}
}

func NotAcceptable(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotAcceptable,
		Code:    CodeNotAcceptable,
		Message: message,
	}
}

func UnsupportedMediaType(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnsupportedMediaType,
		Code:    CodeUnsupportedMedia,
		Message: message,
	}
}

func TooManyRequests(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusTooManyRequests,
		Code:    CodeTooManyRequests,
		Message: message,
	}
}

// Internal sengaja memakai pesan yang seragam dan tidak informatif.
// Detail teknisnya disimpan pada cause dan hanya muncul di log, karena
// pesan error database sering membocorkan nama tabel dan struktur query.
func Internal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "terjadi kesalahan pada server",
		cause:   cause,
	}
}
