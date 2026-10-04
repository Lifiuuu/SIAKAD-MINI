CREATE TABLE IF NOT EXISTS enrollments (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(20) NOT NULL, -- e.g. "2026/2027-Ganjil"
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Mahasiswa tidak dapat mengambil mata kuliah yang sama dua kali pada tahun akademik yang sama.
    UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS enrollments_student_id_idx
    ON enrollments (student_id);

CREATE INDEX IF NOT EXISTS enrollments_course_id_idx
    ON enrollments (course_id);
