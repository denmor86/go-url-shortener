-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx
ON URLs(original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx;
-- +goose StatementEnd
