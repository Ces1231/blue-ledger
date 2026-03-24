-- Migration 003: Seed default point economy, badges, and quests
-- These are chapter-level defaults. When a new chapter is created,
-- the onboarding flow should call a stored procedure or API endpoint
-- to copy these defaults into the new chapter's rows.
-- This migration is a reference/documentation of default values.

-- NOTE: These are example rows using a placeholder UUID for the founding chapter.
-- In production, use the chapter creation API which seeds these per-chapter.

-- Default Point Economy (matching prototype POINT_ECONOMY constant):
-- INSERT INTO point_economy (chapter_id, activity, xp, category) VALUES
--   ('<chapter-id>', 'Chapter Meeting (on time)', 100, 'Attendance'),
--   ('<chapter-id>', 'Chapter Meeting (late)', 50, 'Attendance'),
--   ('<chapter-id>', 'Service Event', 75, 'Service'),
--   ('<chapter-id>', 'Props Received', 10, 'Brotherhood'),
--   ('<chapter-id>', 'Dues Paid', 100, 'Financial'),
--   ('<chapter-id>', 'Committee Meeting', 50, 'Leadership'),
--   ('<chapter-id>', 'RSVP to Event', 10, 'Engagement'),
--   ('<chapter-id>', 'Complete Quest Step', 25, 'Achievement'),
--   ('<chapter-id>', 'Social Event', 50, 'Social'),
--   ('<chapter-id>', 'Badge Earned (bonus)', 25, 'Achievement');

-- Default Badges (matching prototype BADGES constant):
-- INSERT INTO badges (chapter_id, name, icon, category, description, xp_reward, rarity, criteria) VALUES
--   ('<chapter-id>', 'Founding Brother',    '🏛️', 'Achievement', 'One of the founding members', 500,  'legendary', '{"type":"admin_award"}'),
--   ('<chapter-id>', 'Punctual',            '⏰', 'Attendance',  'Attended 10 meetings on time', 100, 'uncommon',  '{"type":"on_time_meetings_gte","value":10}'),
--   ('<chapter-id>', 'Service Champion',    '🤝', 'Service',     '25+ verified service hours', 200,   'rare',      '{"type":"service_hours_gte","value":25}'),
--   ('<chapter-id>', 'Faithful',            '💎', 'Financial',   'Paid dues on time 3+ semesters', 150,'uncommon',  '{"type":"dues_on_time_consecutive_semesters_gte","value":3}'),
--   ('<chapter-id>', 'Props King',          '👑', 'Brotherhood', 'Received 50+ props', 100,           'rare',      '{"type":"props_received_gte","value":50}'),
--   ('<chapter-id>', 'Chapter Legend',      '⭐', 'Achievement', 'Reached Chapter Icon level (2500+ XP)', 250,'legendary','{"type":"xp_gte","value":2500}'),
--   ('<chapter-id>', '25-Hour Builder',     '🔨', 'Service',     '25 verified service hours', 100,   'common',    '{"type":"service_hours_gte","value":25}'),
--   ('<chapter-id>', 'Rising Star',         '🌟', 'Achievement', 'First 100 XP earned', 50,          'common',    '{"type":"xp_gte","value":100}'),
--   ('<chapter-id>', 'Go-Getter',           '🚀', 'Achievement', 'First event check-in', 25,         'common',    '{"type":"checkin_count_gte","value":1}'),
--   ('<chapter-id>', 'Brotherhood Bond',    '🤲', 'Brotherhood', 'Given 10+ props to brothers', 75,  'uncommon',  '{"type":"props_given_gte","value":10}');

-- This file intentionally left as comments — see cmd/seed/main.go for the seed script.
SELECT 1; -- no-op to make migration valid SQL
