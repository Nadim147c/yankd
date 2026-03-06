-- +goose Up
-- +goose StatementBegin
CREATE VIRTUAL TABLE IF NOT EXISTS contents_fts USING fts5 (text, content = 'contents', content_rowid = 'id');

CREATE TRIGGER IF NOT EXISTS contents_fts_insert AFTER INSERT ON contents BEGIN
INSERT INTO
  contents_fts (rowid, text)
VALUES
  (new.id, new.text);

END;

CREATE TRIGGER IF NOT EXISTS contents_fts_delete AFTER DELETE ON contents BEGIN
INSERT INTO
  contents_fts (contents_fts, rowid, text)
VALUES
  ('delete', old.id, old.text);

END;

CREATE TRIGGER IF NOT EXISTS contents_fts_update AFTER
UPDATE ON contents BEGIN
INSERT INTO
  contents_fts (contents_fts, rowid, text)
VALUES
  ('delete', old.id, old.text);

INSERT INTO
  contents_fts (rowid, text)
VALUES
  (new.id, new.text);

END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER blobs_ai;
DROP TRIGGER blobs_ad;
DROP TRIGGER blobs_au;
DROP TABLE blobs_fts;
-- +goose StatementEnd
