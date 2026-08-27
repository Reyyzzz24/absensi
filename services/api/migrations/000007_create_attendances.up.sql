-- Maps legacy `presensi` table. Unifies the two legacy parallel flows
-- (PresensiController + AbsensiController/EOS) into one -- see D-8.
-- Idempotency/race-condition fix per D-9: UNIQUE(employee_id, work_date) replaces
-- the legacy check-then-act pattern; application layer wraps check-in/out in a
-- transaction with row lock (SELECT ... FOR UPDATE) on top of this constraint.
-- work_date is the shift's start date -- for overnight shifts (is_overnight per
-- shifts.is_overnight, D-14) the checkout after midnight still updates this same
-- row rather than creating a new one for the following calendar day.

CREATE TYPE attendance_status AS ENUM ('open', 'closed', 'flagged_no_checkout');

CREATE TABLE attendances (
    id                  BIGSERIAL PRIMARY KEY,
    employee_id         BIGINT NOT NULL REFERENCES employees(id),
    work_date           DATE NOT NULL,
    shift_id            BIGINT REFERENCES shifts(id),
    is_wfh              BOOLEAN NOT NULL DEFAULT false,

    office_location_id  BIGINT REFERENCES office_locations(id),

    check_in_at         TIMESTAMPTZ,
    check_in_lat        DOUBLE PRECISION,
    check_in_lng        DOUBLE PRECISION,
    check_in_distance_m INTEGER, -- computed distance to office_location at check-in
    check_in_photo_path VARCHAR(255),
    is_late             BOOLEAN, -- vs shift.start_time + shift.late_grace_minutes (D-10)

    check_out_at         TIMESTAMPTZ,
    check_out_lat         DOUBLE PRECISION,
    check_out_lng         DOUBLE PRECISION,
    check_out_distance_m  INTEGER,
    check_out_photo_path  VARCHAR(255),
    is_early_leave        BOOLEAN, -- pulang cepat flag, D-13 (no full overtime calc)

    status              attendance_status NOT NULL DEFAULT 'open',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (employee_id, work_date)
);

CREATE INDEX idx_attendances_employee_date ON attendances(employee_id, work_date);
CREATE INDEX idx_attendances_status ON attendances(status) WHERE status = 'open';
