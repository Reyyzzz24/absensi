-- Maps legacy `karyawan` table. Self-service login guard.
CREATE TABLE employees (
    id            BIGSERIAL PRIMARY KEY,
    nik           VARCHAR(30) NOT NULL UNIQUE, -- legacy nik, used as login identifier
    full_name     VARCHAR(150) NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt, portable from legacy hashes
    department_id BIGINT REFERENCES departments(id),
    position      VARCHAR(100), -- jabatan; drives default shift assignment per D-10
    phone         VARCHAR(30),
    photo_path    VARCHAR(255),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_employees_department_id ON employees(department_id);
