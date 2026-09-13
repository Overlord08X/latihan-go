package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository interface {
	Save(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	// FindAndRevoke mencari token aktif berdasarkan hash-nya, kemudian langsung
	// menandainya sebagai revoked (rotasi). Mengembalikan userID pemilik token.
	FindAndRevoke(ctx context.Context, tokenHash string) (userID int, err error)
	// RevokeAllByUserID mencabut seluruh token aktif milik user tertentu.
	// Dipanggil saat logout total atau saat terdeteksi token reuse.
	RevokeAllByUserID(ctx context.Context, userID int) error
}

type pgTokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) TokenRepository {
	return &pgTokenRepository{db: db}
}

func (r *pgTokenRepository) Save(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *pgTokenRepository) FindAndRevoke(ctx context.Context, tokenHash string) (int, error) {
	var userID int
	var isRevoked bool
	var expiresAt time.Time

	err := r.db.QueryRow(ctx,
		`SELECT user_id, is_revoked, expires_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &isRevoked, &expiresAt)

	if err != nil {
		return 0, ErrNotFound
	}

	if isRevoked {
		// Token sudah dipakai sebelumnya — kemungkinan pencurian token.
		// Cabut seluruh token milik user ini sebagai respons keamanan.
		_ = r.RevokeAllByUserID(ctx, userID)
		return 0, ErrNotFound
	}

	if time.Now().After(expiresAt) {
		return 0, ErrNotFound
	}

	// Rotasi: tandai token lama sebagai revoked
	_, err = r.db.Exec(ctx,
		`UPDATE refresh_tokens SET is_revoked = TRUE WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *pgTokenRepository) RevokeAllByUserID(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = $1 AND is_revoked = FALSE`,
		userID,
	)
	return err
}
