CREATE TABLE IF NOT EXISTS nilais (
    id_nilai    SERIAL          PRIMARY KEY,
    id_student  INT             NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    nama_matkul VARCHAR(255)    NOT NULL,
    nilai       NUMERIC(5, 2)   NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Index pada id_student untuk pencarian cepat berdasarkan mahasiswa
CREATE INDEX IF NOT EXISTS nilais_id_student_idx ON nilais(id_student);
