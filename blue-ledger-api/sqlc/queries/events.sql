-- name: GetEvent :one
SELECT id, chapter_id, name, description, event_date, event_time, location,
       event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, is_active,
       qr_code_token, created_by, created_at, updated_at
FROM events
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListEvents :many
SELECT id, chapter_id, name, description, event_date, event_time, location,
       event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, is_active,
       qr_code_token, created_by, created_at, updated_at
FROM events
WHERE chapter_id = $1 AND deleted_at IS NULL
ORDER BY event_date ASC
LIMIT $2 OFFSET $3;

-- name: CreateEvent :one
INSERT INTO events (chapter_id, name, description, event_date, event_time,
                    location, event_type, xp_attend, xp_rsvp, rsvp_deadline,
                    capacity, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, chapter_id, name, description, event_date, event_time, location,
          event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, is_active,
          qr_code_token, created_by, created_at, updated_at;

-- name: UpdateEvent :one
UPDATE events
SET name          = COALESCE(NULLIF($2,''), name),
    description   = COALESCE($3, description),
    event_date    = COALESCE(NULLIF($4::text,'')::date, event_date),
    event_time    = COALESCE($5, event_time),
    location      = COALESCE($6, location),
    event_type    = COALESCE(NULLIF($7,''), event_type),
    xp_attend     = CASE WHEN $8 > 0 THEN $8 ELSE xp_attend END,
    xp_rsvp       = CASE WHEN $9 > 0 THEN $9 ELSE xp_rsvp END,
    rsvp_deadline = COALESCE($10, rsvp_deadline),
    capacity      = COALESCE($11, capacity),
    updated_at    = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, chapter_id, name, event_date, event_type, xp_attend, updated_at;

-- name: CheckInMember :exec
INSERT INTO rsvps (chapter_id, event_id, member_id, status, attended, checked_in_at, checked_in_by)
VALUES ($1, $2, $3, 'yes', true, NOW(), $4)
ON CONFLICT (event_id, member_id)
DO UPDATE SET attended = true, checked_in_at = NOW(), checked_in_by = $4, updated_at = NOW();

-- name: GetEventAttendance :many
SELECT r.id, r.member_id, r.status, r.attended, r.checked_in_at,
       u.first_name, u.last_name, m.member_display_id
FROM rsvps r
JOIN members m ON m.id = r.member_id
JOIN users u ON u.id = m.user_id
WHERE r.event_id = $1
ORDER BY r.attended DESC, r.created_at ASC;
