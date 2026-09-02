CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT students_nim_unique UNIQUE (nim),
    CONSTRAINT students_grade_check CHECK (grade >= 0 AND grade <= 100)
);

CREATE INDEX IF NOT EXISTS students_name_idx
ON students (name);