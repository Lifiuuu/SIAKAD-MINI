CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    nim VARCHAR(12) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    prodi VARCHAR(100) NOT NULL,
    angkatan SMALLINT NOT NULL,
    ipk_terakhir DOUBLE PRECISION NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id) -- 1-1 relationship with users
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

CREATE INDEX IF NOT EXISTS students_nama_lower_idx
    ON students (LOWER(nama));

CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);
