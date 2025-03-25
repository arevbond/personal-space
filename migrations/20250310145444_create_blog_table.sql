-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title TEXT,
    description TEXT,
    body TEXT,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW ()
);

CREATE TABLE IF NOT EXISTS cv (
    id SERIAL PRIMARY KEY,
    name TEXT,
    content BYTEA NOT NULL,
    file_extension TEXT,
    last_updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        version INTEGER DEFAULT 1
);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS cv;
