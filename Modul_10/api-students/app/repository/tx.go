package repository

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager adalah satu-satunya tempat transaction dibuka dan ditutup.
type TxManager struct {
	pool        *pgxpool.Pool
	maxAttempts int
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool, maxAttempts: 3}
}

// WithTx menjalankan fn di dalam satu transaction.
//
// fn menerima DBTX, bukan pgx.Tx, supaya layer service tidak perlu
// mengimpor pgx sama sekali.
//
// Bila fn mengembalikan error, seluruh perubahan dibatalkan. Bila fn
// selesai tanpa error, seluruh perubahan disimpan sekaligus.
//
// Kegagalan karena deadlock (40P01) atau serialization failure (40001)
// dicoba ulang otomatis. Konsekuensinya: fn BISA dijalankan lebih dari
// sekali. Jangan pernah melakukan sesuatu di luar database di dalam fn  - 
// mengirim email, memanggil API lain  -  karena tindakan itu tidak ikut
// dibatalkan dan bisa terjadi berulang kali.
func (m *TxManager) WithTx(ctx context.Context, fn func(tx DBTX) error) error {
	var err error
	for attempt := 1; attempt <= m.maxAttempts; attempt++ {
		err = m.runOnce(ctx, fn)
		if !isRetryable(err) {
			return err
		}
		// Jeda singkat yang sedikit acak, agar dua transaction yang baru
		// saja bertabrakan tidak langsung bertabrakan lagi di detik yang sama.
		wait := time.Duration(attempt*20+rand.IntN(30)) * time.Millisecond
		select {
		case <-ctx.Done():
			return err
		case <-time.After(wait):
		}
	}
	return err
}

func (m *TxManager) runOnce(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaction: %w", err)
	}

	// Rollback SELALU dijadwalkan. Setelah Commit berhasil, Rollback tidak
	// melakukan apa-apa (pgx mengembalikan ErrTxClosed yang diabaikan).
	// Tanpa baris ini, setiap kegagalan meninggalkan transaction terbuka
	// yang terus memegang satu koneksi dari pool  -  selamanya.
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func isRetryable(err error) bool {
	code := pgErrorCode(err)
	return code == codeSerializationFailure || code == codeDeadlockDetected
}
