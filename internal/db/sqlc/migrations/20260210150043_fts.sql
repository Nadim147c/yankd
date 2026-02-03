-- +goose Up
-- +goose StatementBegin
CREATE VIRTUAL TABLE IF NOT EXISTS datas_fts USING fts5 (text, content = 'datas', content_rowid = 'id');

CREATE TRIGGER IF NOT EXISTS datas_fts_insert AFTER INSERT ON datas BEGIN
INSERT INTO
  datas_fts (rowid, text)
VALUES
  (new.id, new.text);

END;

CREATE TRIGGER IF NOT EXISTS datas_fts_delete AFTER DELETE ON datas BEGIN
INSERT INTO
  datas_fts (datas_fts, rowid, text)
VALUES
  ('delete', old.id, old.text);

END;

CREATE TRIGGER IF NOT EXISTS datas_fts_update AFTER
UPDATE ON datas BEGIN
INSERT INTO
  datas_fts (datas_fts, rowid, text)
VALUES
  ('delete', old.id, old.text);

INSERT INTO
  datas_fts (rowid, text)
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
