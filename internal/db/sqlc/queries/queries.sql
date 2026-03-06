-- name: CreateEvent :one
INSERT INTO events (primary_mime_type, time)
VALUES (?, ?)
RETURNING *;

-- name: CreateContent :one
INSERT INTO contents (hash, is_text, text, blob)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreateEntry :one
INSERT INTO entries (event_id, mime_type, content_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetEvents :many
SELECT * FROM events
WHERE id IN (sqlc.slice('ids'));

-- name: GetLastEvents :many
SELECT *
FROM events
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

-- name: FullTextSearch :many
WITH matched_contents AS (
  SELECT text FROM contents_fts
  WHERE contents_fts.text MATCH ?
)
SELECT e.id, MIN(m.rank) AS best_rank
FROM matched_contents m
JOIN contents d ON d.id = m.rowid
JOIN entries en ON en.content_id = d.id
JOIN events e ON e.id = en.event_id
GROUP BY e.id
ORDER BY best_rank;
