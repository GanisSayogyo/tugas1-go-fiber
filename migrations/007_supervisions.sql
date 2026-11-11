
-- Migration 007: Supervisors dan Supervisions

CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT supervisors_count_check
        CHECK (
            supervised_count >= 0
            AND supervised_count <= quota
        )
);

CREATE TABLE IF NOT EXISTS supervisions (
    student_id INTEGER PRIMARY KEY
        REFERENCES students(id) ON DELETE CASCADE,

    supervisor_id INTEGER NOT NULL
        REFERENCES supervisors(id) ON DELETE RESTRICT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx
ON supervisions (supervisor_id);

-- Permission untuk mengelola pembimbing tugas akhir
INSERT INTO permissions (name, description)
VALUES (
    'supervision:manage',
    'Mengelola penetapan dosen pembimbing tugas akhir'
)
ON CONFLICT (name) DO NOTHING;

-- Hanya admin yang mendapat permission ini
INSERT INTO role_permissions (role_name, permission_name)
VALUES ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;
