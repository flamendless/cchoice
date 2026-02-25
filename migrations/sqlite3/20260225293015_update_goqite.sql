-- +goose Up
-- +goose StatementBegin
ALTER TABLE goqite ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
CREATE INDEX goqite_queue_priority_created_idx on goqite (queue, priority desc, created);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX goqite_queue_priority_created_idx;
-- +goose StatementEnd
