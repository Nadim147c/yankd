-- name: CreateEvent :one
INSERT INTO events (primary_mime_type, time)
VALUES (?, ?)
RETURNING *;

-- name: CreateData :one
INSERT INTO datas (hash, is_text, text, blob)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreateEntry :one
INSERT INTO entries (event_id, mime_type, data_id)
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
SELECT * FROM datas
WHERE hash = ?;

-- name: GetEntries :many
SELECT entries.*, datas.* FROM entries
JOIN datas ON datas.id = entries.data_id
WHERE entries.event_id IN (sqlc.slice('ids'));

-- name: FullTextSearch :many
WITH matched_datas AS (
  SELECT text FROM datas_fts
  WHERE datas_fts.text MATCH ?
)
SELECT e.id, MIN(m.rank) AS best_rank
FROM matched_datas m
JOIN datas d ON d.id = m.rowid
JOIN entries en ON en.data_id = d.id
JOIN events e ON e.id = en.event_id
GROUP BY e.id
ORDER BY best_rank;
