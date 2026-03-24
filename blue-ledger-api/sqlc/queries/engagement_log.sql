-- name: InsertEngagementLog :one
INSERT INTO engagement_log (chapter_id, member_id, activity, xp_awarded, source,
                             reference_id, reference_type, awarded_by, note, semester)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, chapter_id, member_id, activity, xp_awarded, source, created_at;

-- name: GetMemberEngagementLog :many
SELECT id, activity, xp_awarded, source, reference_id, reference_type, semester, note, created_at
FROM engagement_log
WHERE member_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SumMemberXP :one
SELECT COALESCE(SUM(xp_awarded), 0)::int AS total
FROM engagement_log
WHERE member_id = $1;

-- name: GetChapterLeaderboard :many
SELECT
    ROW_NUMBER() OVER (ORDER BY SUM(el.xp_awarded) DESC) AS rank,
    el.member_id,
    SUM(el.xp_awarded)::int AS total_xp,
    u.first_name,
    u.last_name,
    m.member_display_id,
    m.level,
    m.level_key
FROM engagement_log el
JOIN members m ON m.id = el.member_id
JOIN users u ON u.id = m.user_id
WHERE el.chapter_id = $1
GROUP BY el.member_id, u.first_name, u.last_name, m.member_display_id, m.level, m.level_key
ORDER BY total_xp DESC;
