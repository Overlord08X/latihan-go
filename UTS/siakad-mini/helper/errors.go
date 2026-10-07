package helper

import (
	"fmt"
)

type AppError struct {
	Status  int
	Message string
	Errors  map[string][]string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(status int, message string) *AppError {
	return &AppError{
		Status:  status,
		Message: message,
	}
}

func NewValidationError(message string, errs map[string][]string) *AppError {
	return &AppError{
		Status:  422,
		Message: message,
		Errors:  errs,
	}
}

func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = "Autentikasi gagal atau token tidak valid"
	}
	return NewError(401, msg)
}

func Forbidden(msg string) *AppError {
	if msg == "" {
		msg = "Anda tidak memiliki izin untuk mengakses resource ini"
	}
	return NewError(403, msg)
}

func NotFound(msg string) *AppError {
	if msg == "" {
		msg = "Data tidak ditemukan"
	}
	return NewError(404, msg)
}

func Conflict(msg string) *AppError {
	return NewError(409, msg)
}

func UnprocessableEntity(msg string, errors map[string][]string) *AppError {
	return NewValidationError(msg, errors)
}

func TooManyRequests(msg string) *AppError {
	if msg == "" {
		msg = "Terlalu banyak permintaan, silakan coba beberapa saat lagi"
	}
	return NewError(429, msg)
}

func Internal(err error, msg string) *AppError {
	if msg == "" {
		msg = "Terjadi kesalahan internal server"
	}
	return &AppError{
		Status:  500,
		Message: msg,
		Err:     err,
	}
}
