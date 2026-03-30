-- Add streak tracking to members table
ALTER TABLE members ADD COLUMN current_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN longest_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN streak_updated_at TIMESTAMP DEFAULT NOW();

-- Create index for streak queries
CREATE INDEX idx_members_current_streak ON members(current_streak DESC);
CREATE INDEX idx_members_longest_streak ON members(longest_streak DESC);
