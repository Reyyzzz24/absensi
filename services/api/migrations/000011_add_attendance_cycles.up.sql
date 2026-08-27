-- Supports multiple check-in/out cycles per employee per day (e.g. legacy GA
-- department security-guard rotations, LOGIC_SPEC.md §3/§5) -- per
-- docs/DECISIONS.md D-23. Replaces the single UNIQUE(employee_id, work_date)
-- constraint from migration 000007 with UNIQUE(employee_id, work_date, cycle_number).

ALTER TABLE attendances DROP CONSTRAINT attendances_employee_id_work_date_key;

ALTER TABLE attendances ADD COLUMN cycle_number INTEGER NOT NULL DEFAULT 1;

ALTER TABLE attendances ADD CONSTRAINT attendances_employee_date_cycle_key
    UNIQUE (employee_id, work_date, cycle_number);
