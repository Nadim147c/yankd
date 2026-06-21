INSTALL fts;

LOAD fts;

CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY DEFAULT uuidv7 (),
  primary_mime_type TEXT NOT NULL DEFAULT 'text/plain',
  time DATETIME NOT NULL DEFAULT get_current_time (),
  preview TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS contents (
  id UUID PRIMARY KEY DEFAULT uuidv7 (),
  hash UHUGEINT NOT NULL,
  is_text BOOLEAN NOT NULL DEFAULT true,
  blob BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS entries (
  event_id UUID NOT NULL,
  mime_type TEXT NOT NULL DEFAULT 'text/plain',
  content_id UUID NOT NULL,
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_events_id ON events (id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_contents_id ON contents (id);

CREATE INDEX IF NOT EXISTS idx_entries_event_id ON entries (event_id);

CREATE INDEX IF NOT EXISTS idx_entries_content_id ON entries (content_id);

CREATE INDEX IF NOT EXISTS idx_contents_hash ON contents (hash);
