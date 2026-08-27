-- Deliberately a single row (id fixed at 1) -- this app has no
-- multi-tenant/workspace concept, so there is exactly one company profile.
CREATE TABLE company_settings (
    id BIGINT PRIMARY KEY DEFAULT 1,
    name TEXT NOT NULL DEFAULT '',
    logo_path TEXT,
    CONSTRAINT company_settings_single_row CHECK (id = 1)
);

INSERT INTO company_settings (id, name) VALUES (1, 'PT Absensi Digital');
