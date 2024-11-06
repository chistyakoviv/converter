-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS conversion_queue
(
    id               BIGSERIAL PRIMARY KEY,
    fullpath         VARCHAR(255) NOT NULL UNIQUE, -- Ensure fullpath is unique so the deletion operation can use 'convert_to' to identify files to delete.
    path             VARCHAR(255) NOT NULL,
    filestem         VARCHAR(255) NOT NULL, -- filename without extension (e. g. /path/to/photo.ext -> photo)
    ext              VARCHAR(255) NOT NULL,
    convert_to       VARCHAR(50)[] NOT NULL,
    is_done          BOOLEAN NOT NULL DEFAULT FALSE,
    is_canceled      BOOLEAN NOT NULL DEFAULT FALSE,
    replace_orig_ext BOOLEAN NOT NULL DEFAULT FALSE,
    error_code       INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP
);
CREATE INDEX IF NOT EXISTS conversion_queue_fullpath_idx ON conversion_queue (fullpath);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS conversion_queue_fullpath_idx;
DROP TABLE IF EXISTS conversion_queue;
-- +goose StatementEnd