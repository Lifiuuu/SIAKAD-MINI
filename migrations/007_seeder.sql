-- Seeder untuk SIAKAD Mini
-- Menambahkan 1 admin, 20 mahasiswa, dan 10 mata kuliah.

-- 1. Insert 1 Admin
INSERT INTO users (email, password, role) VALUES 
('admin@siakad.ac.id', '$2a$12$XEm1zRYkXfziXAssxcwRxORL2yQs7q3DX/aeU29oAmmhzoZkp.jnm', 'admin') -- contoh password hash 'password'
ON CONFLICT DO NOTHING;

-- 2. Insert 20 Mahasiswa
DO $$
DECLARE
    i INT;
    uid INT;
BEGIN
    FOR i IN 1..20 LOOP
        INSERT INTO users (email, password, role) 
        VALUES ('student' || i || '@siakad.ac.id', '$2a$12$XEm1zRYkXfziXAssxcwRxORL2yQs7q3DX/aeU29oAmmhzoZkp.jnm', 'mahasiswa')
        RETURNING id INTO uid;

        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
        VALUES (
            uid, 
            '18722100' || LPAD(i::text, 4, '0'),
            'Mahasiswa ' || i,
            'Informatika',
            2026,
            (random() * (4.0 - 2.0) + 2.0) -- Random IPK between 2.0 and 4.0
        );
    END LOOP;
END $$;

-- 3. Insert 10 Mata Kuliah
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES 
('IF101', 'Advanced Backend Programming', 3, 1, 40),
('IF102', 'Machine Learning', 3, 2, 40),
('IF103', 'Project 1', 3, 3, 40),
('IF104', 'Islamic Studies II', 3, 4, 40),
('IF105', 'Design Thinking', 3, 5, 40),
('IF106', 'Quality Assurance', 3, 4, 40),
('IF107', 'Kecerdasan Buatan', 3, 6, 40),
('IF108', 'Rekayasa Perangkat Lunak', 3, 5, 40),
('IF109', 'IT Entrepreneurship', 3, 7, 40),
('IF110', 'Cyber Security', 3, 7, 40)
ON CONFLICT DO NOTHING;
