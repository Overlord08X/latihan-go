package service

import (
	"errors"
	"strings"
)

// Daftar password yang terlalu umum — ditolak meskipun memenuhi panjang minimum.
// Ini mencegah password seperti "password1" yang mudah ditebak.
var commonPasswords = map[string]struct{}{
	"password":   {},
	"password1":  {},
	"password123":{},
	"12345678":   {},
	"123456789":  {},
	"1234567890": {},
	"qwerty123":  {},
	"iloveyou":   {},
	"admin123":   {},
	"letmein":    {},
	"welcome1":   {},
	"monkey123":  {},
}

// ValidatePasswordStrength memeriksa kekuatan password secara murni (tanpa Fiber).
// Mengembalikan error bila:
//   - kurang dari 8 karakter
//   - termasuk dalam daftar password umum
//
// Fungsi ini tidak melakukan sanitasi atau perubahan — hanya membaca dan memvalidasi.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password minimal 8 karakter")
	}

	lowered := strings.ToLower(strings.TrimSpace(password))
	if _, found := commonPasswords[lowered]; found {
		return errors.New("password terlalu umum, gunakan kombinasi yang lebih unik")
	}

	return nil
}

// ValidateRegister memvalidasi seluruh input pendaftaran.
// Mengembalikan map error per field dan error umum jika ada kegagalan.
func ValidateRegister(username, email, password string) (map[string]string, error) {
	errs := make(map[string]string)

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" {
		errs["username"] = "username wajib diisi"
	} else if len(username) < 3 {
		errs["username"] = "username minimal 3 karakter"
	}

	if email == "" {
		errs["email"] = "email wajib diisi"
	} else if !strings.Contains(email, "@") {
		errs["email"] = "email tidak valid"
	}

	if password == "" {
		errs["password"] = "password wajib diisi"
	} else if err := ValidatePasswordStrength(password); err != nil {
		errs["password"] = err.Error()
	}

	if len(errs) > 0 {
		return errs, errors.New("validasi gagal")
	}

	return nil, nil
}
