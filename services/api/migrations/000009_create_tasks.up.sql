-- Maps legacy `tasks` table (the only domain table that had a real Laravel
-- migration: database/migrations/2026_08_12_105139_create_tasks_table.php).
-- Legacy IDOR: edit/update endpoints did not scope by ownership (A5/D-5) --
-- fixed at the application layer (handler must filter by employee_id = current user).
CREATE TABLE tasks (
    id           BIGSERIAL PRIMARY KEY,
    employee_id  BIGINT NOT NULL REFERENCES employees(id),
    title        VARCHAR(255) NOT NULL,      -- nama_pekerjaan
    detail       TEXT,                       -- detail_pekerjaan
    starts_at    TIMESTAMPTZ NOT NULL,        -- jam_mulai
    ends_at      TIMESTAMPTZ,                 -- jam_selesai
    status       VARCHAR(30) NOT NULL DEFAULT 'planned',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_employee_id ON tasks(employee_id);
