-- name: CreateEvent :one
INSERT INTO events (primary_mime_type, time, preview)
VALUES (?, ?, ?)
RETURNING *;

-- name: CreateContent :one
INSERT INTO contents (hash, is_text, blob)
VALUES (?, ?, ?)
RETURNING *;

-- name: CreateEntry :one
INSERT INTO entries (event_id, mime_type, content_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetEvents :many
SELECT * FROM events
WHERE events.id IN (sqlc.slice('ids'));

-- name: GetLastEvents :many
SELECT * FROM events
ORDER BY time DESC
LIMIT ?;

-- name: DeleteEvents :execrows
DELETE FROM events
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteAllEvents :execrows
DELETE FROM events;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = ?;

-- name: GetDatasByHash :many
SELECT * FROM contents
WHERE hash = ?;

-- name: GetEntries :many
SELECT entries.*, contents.* FROM entries
JOIN contents ON contents.id = entries.content_id
WHERE entries.event_id IN (sqlc.slice('ids'));

-- name: GetEventsPreviewAndID :many
SELECT id, primary_mime_type, preview FROM events
ORDER BY time DESC
LIMIT ?;