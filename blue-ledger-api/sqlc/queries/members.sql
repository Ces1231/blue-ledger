-- name: GetMember :one
SELECT m.id, m.chapter_id, m.user_id, m.member_display_id,
       u.first_name, u.last_name, u.display_name, u.email, u.avatar_url,
       m.role, m.status, m.inducted_year, m.employer, m.job_title,
       m.city, m.linkedin_url, m.xp_total, m.xp_semester,
       m.level, m.level_key, m.dues_status, m.avatar_bg, m.avatar_fg,
       m.created_at, m.updated_at
FROM members m
JOIN users u ON u.id = m.user_id
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: GetMemberByDisplayID :one
SELECT m.id, m.chapter_id, m.user_id, m.member_display_id,
       u.first_name, u.last_name, u.display_name, u.email, u.avatar_url,
       m.role, m.status, m.inducted_year, m.employer, m.job_title,
       m.city, m.linkedin_url, m.xp_total, m.xp_semester,
       m.level, m.level_key, m.dues_status, m.avatar_bg, m.avatar_fg,
       m.created_at, m.updated_at
FROM members m
JOIN users u ON u.id = m.user_id
WHERE m.chapter_id = $1 AND m.member_display_id = $2 AND m.deleted_at IS NULL;

-- name: GetMemberByEmail :one
SELECT m.id, m.chapter_id, m.user_id, m.member_display_id,
       u.first_name, u.last_name, u.display_name, u.email, u.avatar_url,
       m.role, m.status, m.xp_total, m.xp_semester, m.level, m.level_key,
       m.dues_status, m.created_at, m.updated_at
FROM members m
JOIN users u ON u.id = m.user_id
WHERE u.email = $1 AND m.chapter_id = $2 AND m.deleted_at IS NULL;

-- name: ListMembers :many
SELECT m.id, m.chapter_id, m.user_id, m.member_display_id,
       u.first_name, u.last_name, u.display_name, u.email, u.avatar_url,
       m.role, m.status, m.inducted_year, m.employer, m.job_title,
       m.city, m.linkedin_url, m.xp_total, m.xp_semester,
       m.level, m.level_key, m.dues_status, m.avatar_bg, m.avatar_fg,
       m.created_at, m.updated_at
FROM members m
JOIN users u ON u.id = m.user_id
WHERE m.chapter_id = $1 AND m.deleted_at IS NULL
ORDER BY m.xp_total DESC
LIMIT $2 OFFSET $3;

-- name: CountMembers :one
SELECT COUNT(*) FROM members WHERE chapter_id = $1 AND deleted_at IS NULL;

-- name: CreateMember :one
INSERT INTO members (chapter_id, user_id, member_display_id, role, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, chapter_id, user_id, member_display_id, role, status, xp_total, xp_semester, level, level_key, dues_status, created_at, updated_at;

-- name: UpdateMember :one
UPDATE members
SET inducted_year = COALESCE($2, inducted_year),
    employer      = COALESCE($3, employer),
    job_title     = COALESCE($4, job_title),
    city          = COALESCE($5, city),
    linkedin_url  = COALESCE($6, linkedin_url),
    avatar_bg     = COALESCE($7, avatar_bg),
    avatar_fg     = COALESCE($8, avatar_fg),
    role          = COALESCE($9, role),
    status        = COALESCE($10, status),
    dues_status   = COALESCE($11, dues_status),
    updated_at    = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, chapter_id, user_id, member_display_id, role, status, xp_total, xp_semester, level, level_key, dues_status, updated_at;

-- name: DeleteMember :exec
UPDATE members SET deleted_at = NOW(), status = 'inactive' WHERE id = $1;

-- name: GetLeaderboard :many
SELECT
    ROW_NUMBER() OVER (ORDER BY m.xp_total DESC) AS rank,
    m.id, m.member_display_id,
    u.first_name, u.last_name, u.avatar_url,
    m.avatar_bg, m.avatar_fg,
    m.xp_total, m.xp_semester, m.level, m.level_key, m.role
FROM members m
JOIN users u ON u.id = m.user_id
WHERE m.chapter_id = $1 AND m.status = 'active' AND m.deleted_at IS NULL
ORDER BY m.xp_total DESC;
