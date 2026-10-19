-- Migration 011: pembimbing tugas akhir untuk entity students (Tugas Mandiri D.2)
CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    student_id INTEGER PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

INSERT INTO permissions (name, description) VALUES
('supervision:manage', 'Menetapkan dan mengelola pembimbing tugas akhir')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;

-- Data awal dosen pembimbing (Dr. Budi Santoso diset kuota 3 untuk uji beban D.3)
INSERT INTO supervisors (name, quota) VALUES
('Dr. Budi Santoso, M.Kom', 3),
('Dr. Siti Aminah, M.T', 10),
('Prof. Hendra Wijaya, Ph.D', 15)
ON CONFLICT DO NOTHING;
