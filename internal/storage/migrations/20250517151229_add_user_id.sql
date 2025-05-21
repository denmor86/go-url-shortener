-- +goose Up
-- +goose StatementBegin
ALTER TABLE URLs
ADD user_uuid TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE URLs
DROP COLUMN user_uuid;
-- +goose StatementEnd
