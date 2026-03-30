-- Create notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'badge', 'level_up', 'streak', 'daily_login', 'challenge', etc.
    title VARCHAR(255) NOT NULL,
    message TEXT,
    action_url VARCHAR(255),
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY(recipient_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for notification queries
CREATE INDEX idx_notifications_recipient_unread ON notifications(recipient_id, read_at)
WHERE read_at IS NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- Add daily login tracking columns to users table
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP;
ALTER TABLE users ADD COLUMN daily_login_streak INT DEFAULT 0;

-- Add login tracking index
CREATE INDEX idx_users_last_login ON users(last_login_at);
