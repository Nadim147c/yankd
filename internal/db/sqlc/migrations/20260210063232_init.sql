-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  primary_mime_type TEXT NOT NULL DEFAULT 'text/plain',
  time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE contents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hash UNSIGNED BIGINT NOT NULL,
  is_text BOOLEAN NOT NULL DEFAULT 1,
  text TEXT,
  blob BLOB
);

CREATE TABLE entries (
  event_id INTEGER NOT NULL,
  mime_type TEXT NOT NULL DEFAULT 'text/plain',
  content_id INTEGER NOT NULL,
  FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
  FOREIGN KEY (content_id) REFERENCES contents (id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_events_id ON events (id);
CREATE INDEX idx_entries_event_id ON entries (event_id);
CREATE INDEX idx_entries_content_id ON entries (content_id);
CREATE UNIQUE INDEX idx_contents_id ON contents (id);
CREATE INDEX idx_contents_hash ON contents (hash);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS contents;
DROP TABLE IF EXISTS events;

-- +goose StatementEnd

