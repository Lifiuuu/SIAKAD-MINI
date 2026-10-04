-- ---------------------------------------------------------------
-- roles — daftar role yang diakui sistem
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh terhadap seluruh data dan pengaturan'),
    ('mahasiswa', 'Hanya boleh melihat profil, mengambil dan membatalkan mata kuliah')
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------
-- permissions — daftar tindakan yang dapat diberikan kepada role
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
    -- Students permissions
    ('students:list', 'Melihat daftar seluruh mahasiswa'),
    ('students:read:any', 'Melihat data mahasiswa mana pun'),
    ('students:create', 'Menambahkan data mahasiswa baru'),
    ('students:update:any', 'Mengubah data mahasiswa mana pun'),
    ('students:delete', 'Menghapus data mahasiswa'),
    -- Courses permissions
    ('courses:list', 'Melihat daftar mata kuliah'),
    ('courses:create', 'Menambahkan mata kuliah baru'),
    ('courses:update:any', 'Mengubah data mata kuliah mana pun'),
    ('courses:delete', 'Menghapus mata kuliah'),
    -- Enrollments permissions
    ('enrollments:list', 'Melihat daftar krs'),
    ('enrollments:read:any', 'Melihat krs mahasiswa mana pun'),
    ('enrollments:create', 'Menambahkan krs'),
    ('enrollments:delete', 'Menghapus krs')
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------
-- role_permissions — tabel penghubung, inti dari model RBAC
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL
        REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL
        REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
    -- Admin: akses penuh
    ('admin', 'students:list'),
    ('admin', 'students:read:any'),
    ('admin', 'students:create'),
    ('admin', 'students:update:any'),
    ('admin', 'students:delete'),
    ('admin', 'courses:list'),
    ('admin', 'courses:create'),
    ('admin', 'courses:update:any'),
    ('admin', 'courses:delete'),
    -- Mahasiswa
    ('mahasiswa', 'courses:list'),
    ('mahasiswa', 'enrollments:list'),
    ('mahasiswa', 'enrollments:create'),
    ('mahasiswa', 'enrollments:delete')
ON CONFLICT DO NOTHING;

-- Kunci column role pada users agar hanya berisi role yang dikenal.
UPDATE users SET role = 'mahasiswa' WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Cursor index untuk user
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
    ON users (created_at DESC, id DESC);
