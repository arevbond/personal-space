-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE posts
ADD COLUMN slug VARCHAR(100) UNIQUE NOT NULL ;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE posts
DROP COLUMN slug;