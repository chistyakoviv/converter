-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS deletion_queue
(
    id               SERIAL PRIMARY KEY,
    fullpath         VARCHAR(255) NOT NULL UNIQUE,
    status           SMALLINT NOT NULL DEFAULT 0,
    media_type       SMALLINT NOT NULL DEFAULT 1,
    error_code       INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP
);
CREATE INDEX IF NOT EXISTS deletion_queue_fullpath_idx ON deletion_queue (fullpath);
CREATE INDEX IF NOT EXISTS deletion_queue_status_media_type_idx ON deletion_queue (status, media_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS deletion_queue_fullpath_idx;
DROP INDEX IF EXISTS deletion_queue_status_media_type_idx;
DROP TABLE IF EXISTS deletion_queue;
-- +goose StatementEnd
