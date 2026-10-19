-- Migrasi 008: Index komposit untuk keyset pagination (cursor pagination)

-- Urutan column pada index HARUS sama persis dengan ORDER BY pada query,
-- termasuk arah DESC-nya. PostgreSQL memakai index ini untuk melompat
-- langsung ke posisi cursor (created_at, id) < ($1, $2).
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
ON users (created_at DESC, id DESC);

-- Index komposit untuk endpoint students (Tugas Mandiri D.3)
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
ON students (created_at DESC, id DESC);
