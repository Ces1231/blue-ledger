-- Remove streak tracking from members table
ALTER TABLE members DROP COLUMN IF EXISTS current_streak;
ALTER TABLE members DROP COLUMN IF EXISTS longest_streak;
ALTER TABLE members DROP COLUMN IF EXISTS streak_updated_at;

-- Drop streak indexes
DROP INDEX IF EXISTS idx_members_current_streak;
DROP INDEX IF EXISTS idx_members_longest_streak;
