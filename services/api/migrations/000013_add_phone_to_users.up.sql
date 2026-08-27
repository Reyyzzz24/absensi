-- Admin users get a phone column matching Employee's, so the shared
-- self-service profile page has the same editable field set on both
-- audiences (foto, no. HP, password).
ALTER TABLE users ADD COLUMN phone VARCHAR(30);
