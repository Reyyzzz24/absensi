-- Holiday management feature: three sources combined by one resolver
-- (weekend / national / company) -- see docs/DECISIONS.md D-25.

-- Configurable work week (per-company), not hardcoded Sat-Sun -- supports
-- 6-day-work-week companies (only Sunday off) etc. ISO weekday numbering
-- (1=Monday..7=Sunday), matching Go's time.Weekday()+iso adjustment used by
-- the resolver.
ALTER TABLE company_settings
    ADD COLUMN working_weekdays SMALLINT[] NOT NULL DEFAULT '{1,2,3,4,5}';

-- National holidays (incl. cuti bersama), cached locally from a sync job --
-- never fetched at request time. `source='manual'` rows were hand-edited by
-- an admin and must never be overwritten by a later sync (sync only
-- upserts rows still at source='sync').
CREATE TABLE national_holidays (
    id              BIGSERIAL PRIMARY KEY,
    holiday_date    DATE NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    year            SMALLINT NOT NULL,
    is_cuti_bersama BOOLEAN NOT NULL DEFAULT false,
    source          TEXT NOT NULL DEFAULT 'sync' CHECK (source IN ('sync', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_national_holidays_year ON national_holidays(year);

-- Company-specific manual holidays (single date or range), admin-managed.
-- Also the mechanism to add a cuti bersama the national sync doesn't have
-- yet, or a company-only closure day.
CREATE TABLE company_holidays (
    id         BIGSERIAL PRIMARY KEY,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'libur' CHECK (type IN ('libur', 'cuti_bersama')),
    note       TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT company_holidays_range_valid CHECK (end_date >= start_date)
);

CREATE INDEX idx_company_holidays_range ON company_holidays(start_date, end_date);

-- Tags an attendance row as having happened on a resolved holiday (any of
-- the three sources) -- e.g. voluntary/overtime work on a libur day.
-- Late-arrival is never computed for these rows (attendance usecase skips
-- it), this column is purely informational for reports/dashboard.
ALTER TABLE attendances ADD COLUMN is_holiday BOOLEAN NOT NULL DEFAULT false;
