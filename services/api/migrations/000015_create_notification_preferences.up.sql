-- Per-user opt-out of specific notification types, surfaced as toggles in
-- Pengaturan > Preferensi Notifikasi. Absence of a row for a given type
-- means "enabled" (opt-out model, not opt-in) -- see notification usecase.
CREATE TABLE notification_preferences (
    recipient_audience TEXT NOT NULL CHECK (recipient_audience IN ('employee', 'admin')),
    recipient_id BIGINT NOT NULL,
    type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (recipient_audience, recipient_id, type)
);
