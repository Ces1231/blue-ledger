-- Drop notifications table and login tracking
DROP TABLE IF EXISTS notifications CASCADE;

-- Remove daily login tracking from users table
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS daily_login_streak;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_last_login;
