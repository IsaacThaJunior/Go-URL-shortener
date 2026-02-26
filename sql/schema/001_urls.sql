-- +goose UP
CREATE TABLE urls (
    id UUID PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- +goose DOWN
DROP TABLE urls;