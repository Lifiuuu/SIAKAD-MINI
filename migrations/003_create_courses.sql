CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    kode_mk VARCHAR(20) NOT NULL,
    nama_mk VARCHAR(255) NOT NULL,
    sks SMALLINT NOT NULL DEFAULT 2,
    semester SMALLINT NOT NULL DEFAULT 1,
    kuota INTEGER NOT NULL DEFAULT 40,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS courses_kode_mk_lower_key
    ON courses (LOWER(kode_mk));

CREATE INDEX IF NOT EXISTS courses_nama_mk_lower_idx
    ON courses (LOWER(nama_mk));
