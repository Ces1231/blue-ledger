-- name: GetUserByEmail :one
SELECT id, email, email_verified, password_hash, first_name, last_name,
       display_name, avatar_url, last_login_at, failed_login_count,
       locked_until, is_sysadmin, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, email_verified, first_name, last_name,
       display_name, avatar_url, last_login_at, is_sysadmin, created_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, email_verified, password_hash, first_name, last_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, first_name, last_name, is_sysadmin, created_at;

-- name: UpdateLoginSuccess :exec
UPDATE users
SET failed_login_count = 0,
    locked_until = NULL,
    last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateLoginFailure :exec
UPDATE users
SET failed_login_count = $2,
    locked_until = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: CreateAuthSession :one
INSERT INTO auth_sessions (user_id, chapter_id, refresh_token, device_hint, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5::inet, $6)
RETURNING id, created_at;

-- name: GetAuthSession :one
SELECT id, user_id, chapter_id, expires_at, revoked_at
FROM auth_sessions
WHERE refresh_token = $1;

-- name: RevokeAuthSession :exec
UPDATE auth_sessions SET revoked_at = NOW() WHERE id = $1;

-- name: CreateMagicLinkToken :exec
INSERT INTO magic_link_tokens (user_id, token_hash, email, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeMagicLinkToken :one
UPDATE magic_link_tokens
SET used_at = NOW()
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
RETURNING id, user_id, email;
