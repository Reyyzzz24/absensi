-- Maps legacy `jam_kerja` table (shift catalog).
-- is_overnight is derived (jam_pulang < jam_masuk), replacing the legacy hardcoded
-- "only SH03 is overnight" behavior -- see docs/DECISIONS.md D-14.
-- late_grace_minutes is admin/superadmin-configurable per shift -- see D-10
-- (legacy hardcoded 09:15/09:00 threshold, inconsistent across screens).
CREATE TABLE shifts (
    id                 BIGSERIAL PRIMARY KEY,
    code               VARCHAR(20) NOT NULL UNIQUE, -- legacy kode_jam_kerja, e.g. 'SH01', 'LBR'
    name               VARCHAR(100) NOT NULL,
    is_day_off         BOOLEAN NOT NULL DEFAULT false, -- replaces legacy magic code 'LBR'
    start_time         TIME, -- jam_masuk; null when is_day_off
    end_time           TIME, -- jam_pulang; null when is_day_off
    is_overnight       BOOLEAN GENERATED ALWAYS AS (
                           start_time IS NOT NULL AND end_time IS NOT NULL AND end_time < start_time
                       ) STORED,
    late_grace_minutes INTEGER NOT NULL DEFAULT 15, -- default 09:15 parity; editable by admin per D-10
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
