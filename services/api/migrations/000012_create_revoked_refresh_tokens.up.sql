-- Denylist for refresh-token revocation on logout. Refresh tokens remain
-- stateless JWTs (no full session store) -- we only record the JTI of
-- tokens explicitly revoked via /auth/logout, checked on /auth/refresh.
-- expires_at mirrors the token's own exp so a cleanup job can purge rows
-- that could no longer be presented anyway (see DECISIONS.md follow-up).
CREATE TABLE revoked_refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    jti TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_revoked_refresh_tokens_expires_at ON revoked_refresh_tokens (expires_at);
