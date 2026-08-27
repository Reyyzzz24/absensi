ALTER TABLE attendances DROP CONSTRAINT attendances_employee_date_cycle_key;
ALTER TABLE attendances DROP COLUMN cycle_number;
ALTER TABLE attendances ADD CONSTRAINT attendances_employee_id_work_date_key UNIQUE (employee_id, work_date);
