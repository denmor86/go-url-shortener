-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS URLs (
   id SERIAL PRIMARY KEY,
   short_url TEXT  NOT NUll,
   original_url TEXT NOT NUll
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE URLs;
-- +goose StatementEnd
