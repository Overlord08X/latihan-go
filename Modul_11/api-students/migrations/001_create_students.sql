-- Migration: 001_create_students
-- Membuat tabel students beserta index yang diperlukan.

CREATE TABLE IF NOT EXISTS students (
    id         SERIAL          PRIMARY KEY,
    nim        VARCHAR(20)     NOT NULL,
    name       VARCHAR(100)    NOT NULL,
    grade      NUMERIC(5, 2)   NOT NULL DEFAULT 0,
    is_active  BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Unique index case-insensitive pada nim agar NIM "2023001" dan "2023001" dianggap sama.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_idx
    ON students (LOWER(nim));

-- Index pada name mempercepat pencarian ILIKE '%search%'.
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
