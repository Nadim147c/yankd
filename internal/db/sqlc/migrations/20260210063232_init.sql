-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  primary_mime_type TEXT NOT NULL DEFAULT 'text/plain',
  time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE datas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hash UNSIGNED BIGINT NOT NULL,
  is_text BOOLEAN NOT NULL DEFAULT 1,
  text TEXT,
  blob BLOB
);

CREATE TABLE entries (
  event_id INTEGER NOT NULL,
  mime_type TEXT NOT NULL DEFAULT 'text/plain',
  data_id INTEGER NOT NULL,
  FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
  FOREIGN KEY (data_id) REFERENCES datas (id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_events_id ON events (id);
CREATE INDEX idx_entries_event_id ON entries (event_id);
CREATE INDEX idx_entries_data_id ON entries (data_id);
CREATE UNIQUE INDEX idx_datas_id ON datas (id);
CREATE INDEX idx_datas_hash ON datas (hash);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS datas;
DROP TABLE IF EXISTS events;

-- +goose StatementEnd

