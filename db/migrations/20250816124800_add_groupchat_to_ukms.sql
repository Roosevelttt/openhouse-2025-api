-- +goose Up
ALTER TABLE ukms ADD COLUMN groupchat VARCHAR(255) NOT NULL;

-- +goose Down
ALTER TABLE ukms DROP COLUMN groupchat;

