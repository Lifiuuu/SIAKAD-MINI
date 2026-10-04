-- Seeder untuk SIAKAD Mini
-- Menambahkan 1 admin, 20 mahasiswa, dan 10 mata kuliah.

-- 1. Insert 1 Admin
INSERT INTO users (email, password, role) VALUES 
('admin@siakad.ac.id', '$2a$10$X8/hV8/O.P7xV6aY.eL/b.ZJb4o5gLpLpLpLpLpLpLpLpLpLpLp', 'admin') -- contoh password hash 'password'
ON CONFLICT DO NOTHING;

-- 2. Insert 20 Mahasiswa
DO $$
DECLARE
    i INT;
    uid INT;
BEGIN
    FOR i IN 1..20 LOOP
        INSERT INTO users (email, password, role) 
        VALUES ('student' || i || '@siakad.ac.id', '$2a$10$X8/hV8/O.P7xV6aY.eL/b.ZJb4o5gLpLpLpLpLpLpLpLpLpLpLp', 'mahasiswa')
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
('IF101', 'Algoritma dan Pemrograman', 3, 1, 40),
('IF102', 'Struktur Data', 3, 2, 40),
('IF103', 'Basis Data', 3, 3, 40),
('IF104', 'Sistem Operasi', 3, 4, 40),
('IF105', 'Jaringan Komputer', 3, 5, 40),
('IF106', 'Pemrograman Web', 3, 4, 40),
('IF107', 'Kecerdasan Buatan', 3, 6, 40),
('IF108', 'Rekayasa Perangkat Lunak', 3, 5, 40),
('IF109', 'Kriptografi', 3, 7, 40),
('IF110', 'Keamanan Siber', 3, 7, 40)
ON CONFLICT DO NOTHING;
