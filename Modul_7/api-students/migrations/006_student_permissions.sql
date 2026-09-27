-- ---------------------------------------------------------------
-- student permissions — hak akses untuk entitas mahasiswa
-- ---------------------------------------------------------------
INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh mahasiswa'),
    ('student:read:any',   'Melihat data mahasiswa mana pun'),
    ('student:create',     'Mendaftarkan data mahasiswa baru'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete',     'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------
-- Pasangkan permissions mahasiswa ke role yang berhak
-- ---------------------------------------------------------------
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------
-- Tambahkan column owner_id pada tabel students.
-- Data lama yang belum memiliki owner_id diisi dengan user pertama
-- sebelum constraint aktif, agar migration tidak gagal.
-- ---------------------------------------------------------------
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER;

-- Mengisi baris-baris lama yang belum punya owner_id dengan user pertama (jika ada user)
DO $$
DECLARE
    first_user_id INTEGER;
BEGIN
    SELECT id INTO first_user_id FROM users ORDER BY id ASC LIMIT 1;
    IF first_user_id IS NOT NULL THEN
        UPDATE students SET owner_id = first_user_id WHERE owner_id IS NULL;
    END IF;
END $$;

-- Tambahkan foreign key constraint
ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_id_fkey;
ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
