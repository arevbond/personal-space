-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title TEXT,
    description TEXT,
    content BYTEA,
    extension VARCHAR,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW ()
);
-- my first table

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

-- +goose StatementEnd
DROP TABLE IF EXISTS posts;
