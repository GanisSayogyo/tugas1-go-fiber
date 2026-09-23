-- Modul 6: memastikan permission student dan ownership tersedia

INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar mahasiswa'),
('student:read:any', 'Melihat mahasiswa mana pun'),
('student:create', 'Membuat data mahasiswa'),
('student:update:any', 'Mengubah mahasiswa mana pun'),
('student:delete', 'Menghapus mahasiswa')
ON CONFLICT (name) DO NOTHING;

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

ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id);