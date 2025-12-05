-- User queries

-- name: CreateUser :one
INSERT INTO users (
    email,
    first_name,
    last_name,
    password_hash
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    email = COALESCE(sqlc.narg('email'), email)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- Game queries

-- name: CreateGame :one
INSERT INTO games (
    owner_id,
    category,
    title,
    description,
    location_name,
    location_address,
    location_point,
    location_notes,
    start_time,
    duration_minutes,
    max_participants,
    pricing_type,
    pricing_amount_cents,
    pricing_currency,
    signup_deadline,
    drop_deadline,
    skill_level,
    notes,
    status
) VALUES (
    sqlc.arg('owner_id'),
    sqlc.arg('category'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('location_name'),
    sqlc.arg('location_address'),
    ST_SetSRID(ST_MakePoint(sqlc.arg('longitude')::float8, sqlc.arg('latitude')::float8), 4326)::geography,
    sqlc.arg('location_notes'),
    sqlc.arg('start_time'),
    sqlc.arg('duration_minutes'),
    sqlc.arg('max_participants'),
    sqlc.arg('pricing_type'),
    sqlc.arg('pricing_amount_cents'),
    sqlc.arg('pricing_currency'),
    sqlc.arg('signup_deadline'),
    sqlc.arg('drop_deadline'),
    sqlc.arg('skill_level'),
    sqlc.arg('notes'),
    sqlc.arg('status')
)
RETURNING id, owner_id, category, title, description, location_name, location_address,
    ST_Y(location_point::geometry) as latitude, ST_X(location_point::geometry) as longitude,
    location_notes, start_time, duration_minutes, max_participants,
    0 as current_participants, 0 as waitlist_count,
    pricing_type, pricing_amount_cents, pricing_currency, signup_deadline,
    drop_deadline, skill_level, notes, status, cancelled_at, created_at, updated_at;

-- name: GetGame :one
SELECT
    g.id, g.owner_id, g.category, g.title, g.description, g.location_name, g.location_address,
    ST_Y(g.location_point::geometry) as latitude, ST_X(g.location_point::geometry) as longitude,
    g.location_notes, g.start_time, g.duration_minutes, g.max_participants,
    COALESCE(COUNT(p.id) FILTER (WHERE p.status = 'confirmed'), 0)::int as current_participants,
    COALESCE(COUNT(p.id) FILTER (WHERE p.status = 'waitlist'), 0)::int as waitlist_count,
    g.pricing_type, g.pricing_amount_cents, g.pricing_currency, g.signup_deadline,
    g.drop_deadline, g.skill_level, g.notes, g.status, g.cancelled_at, g.created_at, g.updated_at
FROM games g
LEFT JOIN participants p ON p.game_id = g.id
WHERE g.id = $1
GROUP BY g.id;

-- name: GetGameForUpdate :one
SELECT
    id, owner_id, category, title, description, location_name, location_address,
    ST_Y(location_point::geometry) as latitude, ST_X(location_point::geometry) as longitude,
    location_notes, start_time, duration_minutes, max_participants,
    pricing_type, pricing_amount_cents, pricing_currency, signup_deadline,
    drop_deadline, skill_level, notes, status, cancelled_at, created_at, updated_at
FROM games
WHERE id = $1
FOR UPDATE;

-- name: ListGamesInRadius :many
SELECT
    g.id, g.owner_id, g.category, g.title, g.description, g.location_name, g.location_address,
    ST_Y(g.location_point::geometry) as latitude, ST_X(g.location_point::geometry) as longitude,
    g.location_notes, g.start_time, g.duration_minutes, g.max_participants,
    COALESCE(COUNT(p.id) FILTER (WHERE p.status = 'confirmed'), 0)::int as current_participants,
    g.pricing_type, g.pricing_amount_cents, g.pricing_currency, g.signup_deadline,
    g.drop_deadline, g.skill_level, g.notes, g.status, g.cancelled_at, g.created_at, g.updated_at,
    up.status as user_participation_status
FROM games g
LEFT JOIN participants p ON p.game_id = g.id AND p.status = 'confirmed'
LEFT JOIN participants up ON up.game_id = g.id AND up.user_id = sqlc.narg('user_id')
WHERE ST_DWithin(
    g.location_point,
    ST_SetSRID(ST_MakePoint(sqlc.arg('longitude')::float8, sqlc.arg('latitude')::float8), 4326)::geography,
    sqlc.arg('radius')::float8
)
AND g.start_time >= sqlc.arg('start_time')
AND (sqlc.narg('end_time')::timestamptz IS NULL OR g.start_time <= sqlc.narg('end_time'))
AND (sqlc.narg('status')::varchar IS NULL OR g.status = sqlc.narg('status'))
AND g.category = ANY(sqlc.arg('categories')::varchar[])
GROUP BY g.id, up.user_id, up.status
ORDER BY g.start_time ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateGame :one
UPDATE games
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    location_name = COALESCE(sqlc.narg('location_name'), location_name),
    location_address = COALESCE(sqlc.narg('location_address'), location_address),
    location_point = CASE
        WHEN sqlc.narg('location_longitude')::float8 IS NOT NULL
            AND sqlc.narg('location_latitude')::float8 IS NOT NULL
        THEN ST_SetSRID(ST_MakePoint(sqlc.narg('location_longitude'), sqlc.narg('location_latitude')), 4326)::geography
        ELSE location_point
    END,
    location_notes = COALESCE(sqlc.narg('location_notes'), location_notes),
    start_time = COALESCE(sqlc.narg('start_time'), start_time),
    duration_minutes = COALESCE(sqlc.narg('duration_minutes'), duration_minutes),
    max_participants = COALESCE(sqlc.narg('max_participants'), max_participants),
    pricing_type = COALESCE(sqlc.narg('pricing_type'), pricing_type),
    pricing_amount_cents = COALESCE(sqlc.narg('pricing_amount_cents'), pricing_amount_cents),
    pricing_currency = COALESCE(sqlc.narg('pricing_currency'), pricing_currency),
    signup_deadline = COALESCE(sqlc.narg('signup_deadline'), signup_deadline),
    drop_deadline = COALESCE(sqlc.narg('drop_deadline'), drop_deadline),
    skill_level = COALESCE(sqlc.narg('skill_level'), skill_level),
    notes = COALESCE(sqlc.narg('notes'), notes),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id;

-- name: CancelGame :one
UPDATE games
SET
    status = 'cancelled',
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id;

-- name: DeleteGame :exec
DELETE FROM games
WHERE id = $1;

-- name: CreateTeam :one
INSERT INTO teams (
    game_id,
    name,
    color
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetTeam :one
SELECT * FROM teams
WHERE id = $1;

-- name: ListTeamsByGame :many
SELECT * FROM teams
WHERE game_id = $1
ORDER BY created_at ASC;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;

-- name: CreateParticipant :one
INSERT INTO participants (
    game_id,
    user_id,
    team_id,
    status,
    paid,
    payment_amount_cents,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetParticipant :one
SELECT * FROM participants
WHERE id = $1;

-- name: GetParticipantByGameAndUser :one
SELECT * FROM participants
WHERE game_id = $1 AND user_id = $2;

-- name: ListParticipantsByGame :many
SELECT
    p.id,
    p.game_id,
    p.user_id,
    p.team_id,
    p.status,
    p.paid,
    p.payment_amount_cents,
    p.notes,
    p.joined_at,
    p.updated_at,
    u.email,
    u.first_name,
    u.last_name
FROM participants p
INNER JOIN users u ON p.user_id = u.id
WHERE p.game_id = $1
ORDER BY p.joined_at ASC;

-- name: ListParticipantsByUser :many
SELECT * FROM participants
WHERE user_id = $1
ORDER BY joined_at DESC;

-- name: UpdateParticipantStatus :one
UPDATE participants
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateParticipantStatusResetJoinedAt :one
UPDATE participants
SET
    status = $2,
    updated_at = NOW(),
    joined_at = NOW()
WHERE id = $1
RETURNING *;

-- name: BatchUpdateParticipantsToConfirmed :exec
UPDATE participants
SET
    status = 'confirmed',
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('participant_ids')::uuid[]);

-- name: BatchUpdateParticipantsToWaitlist :exec
UPDATE participants
SET
    status = 'waitlist',
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('participant_ids')::uuid[]);

-- name: UpdateParticipantPayment :one
UPDATE participants
SET
    paid = $2,
    payment_amount_cents = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateParticipantTeam :one
UPDATE participants
SET
    team_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteParticipant :exec
DELETE FROM participants
WHERE id = $1;

-- name: CountConfirmedParticipants :one
SELECT COUNT(*) FROM participants
WHERE game_id = $1 AND status = 'confirmed';

-- name: CountWaitlistParticipants :one
SELECT COUNT(*) FROM participants
WHERE game_id = $1 AND status = 'waitlist';

-- name: ListParticipantsByGames :many
SELECT
    p.id,
    p.game_id,
    p.user_id,
    p.team_id,
    p.status,
    p.paid,
    p.payment_amount_cents,
    p.notes,
    p.joined_at,
    p.updated_at,
    u.email,
    u.first_name,
    u.last_name
FROM participants p
INNER JOIN users u ON p.user_id = u.id
WHERE p.game_id = ANY(sqlc.arg('game_ids')::uuid[])
ORDER BY p.game_id, p.joined_at ASC;
