CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Akses penuh'),
('staff', 'Akses data mahasiswa'),
('user', 'Pengguna biasa')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar mahasiswa'),
('student:read:any', 'Melihat mahasiswa mana pun'),
('student:create', 'Membuat data mahasiswa'),
('student:update:any', 'Mengubah mahasiswa mana pun'),
('student:delete', 'Menghapus mahasiswa')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

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

UPDATE users
SET role = 'user'
WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_fkey;

ALTER TABLE users
ADD CONSTRAINT users_role_fkey
FOREIGN KEY (role) REFERENCES roles(name);

ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id);