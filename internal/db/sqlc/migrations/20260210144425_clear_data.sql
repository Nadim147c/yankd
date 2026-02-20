-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS delete_orphan_contents AFTER DELETE ON entries BEGIN
DELETE FROM contents
WHERE id = OLD.content_id
  AND NOT EXISTS (SELECT 1 FROM entries WHERE content_id = OLD.content_id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE TRIGGER delete_orphan_contents;
-- +goose StatementEnd
