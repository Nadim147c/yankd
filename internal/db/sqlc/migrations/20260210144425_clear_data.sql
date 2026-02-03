-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS delete_orphan_datas AFTER DELETE ON entries BEGIN
DELETE FROM datas
WHERE id = OLD.data_id
  AND NOT EXISTS (SELECT 1 FROM entries WHERE data_id = OLD.data_id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE TRIGGER delete_orphan_datas;
-- +goose StatementEnd
