-- +goose Up
-- +goose statementBegin
CREATE INDEX IF NOT EXISTS movies_title_idx ON movies USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS movies_genres_idx ON movies USING GIN (genres);

-- +goose statementEnd

-- +goose Down
-- +goose statementBegin
DROP INDEX IF EXISTS movies_title_idx;
DROP INDEX IF EXISTS movies_genres_idx;

-- +goose statementEnd
