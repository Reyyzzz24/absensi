-- Maps legacy `departemen` table.
CREATE TABLE departments (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(20) NOT NULL UNIQUE, -- legacy kode_dept, e.g. 'GA'
    name        VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
