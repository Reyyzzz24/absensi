-- Maps legacy `pengajuan_izin`. Legacy had a dead-end "buat sakit" form with no
-- store handler; merged into a single form with a `type` field per D-16.
-- Approval workflow added per D-15 (legacy auto-accepted with no review).
-- Approved leave excludes the covered dates from "absent" in attendance reports
-- (enforced in report/rekap query logic, not in this table).

CREATE TYPE leave_type AS ENUM ('izin', 'sakit');
CREATE TYPE leave_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE leave_requests (
    id           BIGSERIAL PRIMARY KEY,
    employee_id  BIGINT NOT NULL REFERENCES employees(id),
    type         leave_type NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NOT NULL,
    reason       TEXT,
    status       leave_status NOT NULL DEFAULT 'pending',
    reviewed_by  BIGINT REFERENCES users(id),
    reviewed_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
