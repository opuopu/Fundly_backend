-- 000002_attachments_user_id.down.sql

DROP INDEX IF EXISTS idx_attachments_transaction_id;
DROP INDEX IF EXISTS idx_attachments_user_id;

ALTER TABLE attachments
    DROP COLUMN IF EXISTS user_id;
