-- Maps legacy `jadwal_kerja` (per-date override) and `konfigurasi_jamkerja`
-- (weekly recurring default). Per docs/DECISIONS.md D-18: work_schedules (per-date)
-- always wins over weekly_shift_defaults when both exist for the same employee+date.

CREATE TABLE weekly_shift_defaults (
    id          BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday
    shift_id    BIGINT NOT NULL REFERENCES shifts(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (employee_id, day_of_week)
);

CREATE TABLE work_schedules (
    id          BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    work_date   DATE NOT NULL,
    shift_id    BIGINT NOT NULL REFERENCES shifts(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (employee_id, work_date)
);

CREATE INDEX idx_work_schedules_employee_date ON work_schedules(employee_id, work_date);
