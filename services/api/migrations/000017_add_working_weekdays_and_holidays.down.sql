ALTER TABLE attendances DROP COLUMN IF EXISTS is_holiday;
DROP TABLE IF EXISTS company_holidays;
DROP TABLE IF EXISTS national_holidays;
ALTER TABLE company_settings DROP COLUMN IF EXISTS working_weekdays;
