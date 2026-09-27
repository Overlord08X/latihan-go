// Package jwtutil menyediakan fungsi-fungsi bantuan untuk JWT.
package jwtutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims adalah klaim tambahan di dalam access token.
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// jwtSecret adalah secret yang dimuat dari environment.
// Aplikasi menolak menyala bila variabel ini kosong (lihat MustLoadJWTSecret).
var jwtSecret []byte

// MustLoadJWTSecret membaca JWT_SECRET dari environment.
// Jika kosong, fungsi ini langsung panic — lebih baik gagal seketika
// daripada berjalan dengan secret kosong yang membuat token bisa dipalsukan siapa saja.
func MustLoadJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		panic("JWT_SECRET harus diisi minimal 32 karakter di environment")
	}
	jwtSecret = []byte(secret)
}

// GenerateAccessToken membuat JWT HS256 dengan masa berlaku sesuai JWT_ACCESS_TTL_MINUTES.
// Claim: sub (userID), username, role, iss, iat, exp.
func GenerateAccessToken(userID int, username, role string, ttlMinutes int) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "praktikum-backend",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlMinutes) * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseAccessToken mem-parse dan memvalidasi JWT.
// Keyfunc secara eksplisit hanya mengizinkan metode HS256.
// Ini menutup kerentanan algorithm confusion (termasuk alg:"none").
func ParseAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Periksa method secara eksplisit — menolak algoritma lain termasuk "none"
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tanda tangan tidak diizinkan: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		// Bedakan token kedaluwarsa dan token tidak valid.
		// Pembedaan ini aman karena bukan tentang rahasia pengguna,
		// melainkan tentang kondisi teknis token itu sendiri.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GenerateRefreshToken membuat string acak 32 byte yang di-encode hex.
// Ini BUKAN JWT — hanya token acak yang disimpan hash-nya di database.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken membuat SHA-256 hex dari token.
// Dipakai untuk menyimpan refresh token di database tanpa menyimpan nilai aslinya.
// SHA-256 sudah cukup di sini (berbeda dengan password) karena token dibuat acak
// 32 byte — tidak ada serangan dictionary yang relevan.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Sentinel errors untuk JWT
var (
	ErrTokenExpired = errors.New("access token kedaluwarsa")
	ErrTokenInvalid = errors.New("access token tidak valid")
)
