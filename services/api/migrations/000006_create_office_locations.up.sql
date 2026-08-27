-- Maps legacy `konfigurasi_lokasi`. Legacy bug: office lat/lng were hardcoded in
-- controller code and this table's radius was computed but never enforced (A1).
-- Per docs/DECISIONS.md D-1: radius IS enforced server-side in the new system,
-- default 100 meters, value read from this table (not hardcoded).
CREATE TABLE office_locations (
    id             BIGSERIAL PRIMARY KEY,
    name           VARCHAR(150) NOT NULL,
    latitude       DOUBLE PRECISION NOT NULL,
    longitude      DOUBLE PRECISION NOT NULL,
    radius_meters  INTEGER NOT NULL DEFAULT 100,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
