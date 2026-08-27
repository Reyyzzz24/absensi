-- "Dinas luar" (legitimate off-site work) support, per docs/DECISIONS.md D-21.
-- Admin pre-approves an employee+date assignment; check-in on that date/employee
-- bypasses the office_locations radius check (D-1) for that day only.
CREATE TABLE field_assignments (
    id           BIGSERIAL PRIMARY KEY,
    employee_id  BIGINT NOT NULL REFERENCES employees(id),
    work_date    DATE NOT NULL,
    note         TEXT,
    approved_by  BIGINT NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (employee_id, work_date)
);

CREATE INDEX idx_field_assignments_employee_date ON field_assignments(employee_id, work_date);
