-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
INSERT INTO categories (id, name)
VALUES (1, 'Книги'), (2, 'Технологии'), (3, 'Разное');

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DELETE FROM categories WHERE id IN (1, 2, 3);
