-- Admin/back-office accounts + roles. RBAC per DECISIONS.md D-7.
-- Provisional: pending review against live DB dump (see docs/DECISIONS.md D-12).

CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL UNIQUE, -- e.g. 'superadmin', 'admin'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(150) UNIQUE,
    username      VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt, portable from legacy Laravel hashes
    role_id       BIGINT NOT NULL REFERENCES roles(id),
    photo_path    VARCHAR(255),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (name) VALUES ('superadmin'), ('admin');
