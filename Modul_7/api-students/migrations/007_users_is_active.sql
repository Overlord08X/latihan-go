-- Migrasi 007: Tambah kolom is_active pada tabel users untuk soft-state / filtering
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
