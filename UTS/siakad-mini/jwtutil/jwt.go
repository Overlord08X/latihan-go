package jwtutil

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token telah kedaluwarsa")
	ErrTokenInvalid = errors.New("token tidak valid")
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID int64  `json:"student_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, email, role string, studentID int64, secret string, ttlHours int) (string, int64, error) {
	if ttlHours <= 0 {
		ttlHours = 24
	}
	expiresIn := int64(ttlHours * 3600)
	now := time.Now()
	exp := now.Add(time.Duration(ttlHours) * time.Hour)

	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		StudentID: studentID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    "siakad-mini",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("gagal menandatangani token: %w", err)
	}

	return signed, expiresIn, nil
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak diharapkan: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
