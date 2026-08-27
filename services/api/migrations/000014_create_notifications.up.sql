-- In-app notifications, MVP-polling model (websocket/push is a future
-- upgrade, not built now). recipient_audience + recipient_id together
-- identify the owner -- an employee id and an admin user id are from
-- separate sequences and can collide numerically, so every query must
-- filter on both columns together, never recipient_id alone.
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    recipient_audience TEXT NOT NULL CHECK (recipient_audience IN ('employee', 'admin')),
    recipient_id BIGINT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    link TEXT,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_recipient ON notifications (recipient_audience, recipient_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications (recipient_audience, recipient_id) WHERE read_at IS NULL;
