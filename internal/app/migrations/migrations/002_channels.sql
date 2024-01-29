-- +goose Up
CREATE TABLE IF NOT EXISTS channels
(
    id INTEGER NOT NULL,
    type TEXT,
    PRIMARY KEY (id)
);

-- +goose Down

DROP TABLE IF EXISTS channels;

