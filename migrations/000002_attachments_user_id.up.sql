-- 000002_attachments_user_id.up.sql
-- Add ownership to attachments so they can be scoped to the uploading user.

ALTER TABLE attachments
    ADD COLUMN IF NOT EXISTS user_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_attachments_user_id ON attachments(user_id);
CREATE INDEX IF NOT EXISTS idx_attachments_transaction_id ON attachments(transaction_id);
