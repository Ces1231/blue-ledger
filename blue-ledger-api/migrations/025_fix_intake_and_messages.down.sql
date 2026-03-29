-- ============================================================
-- Rollback: 025_fix_intake_and_messages
-- ============================================================

DROP TABLE IF EXISTS thread_participants;

ALTER TABLE message_threads
    DROP COLUMN IF EXISTS subject,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS is_group_chat;

ALTER TABLE intake_prospects
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS university,
    DROP COLUMN IF EXISTS grad_year,
    DROP COLUMN IF EXISTS added_by,
    DROP COLUMN IF EXISTS archived_at;
