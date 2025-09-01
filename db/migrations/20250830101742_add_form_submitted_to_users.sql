-- +goose Up
ALTER TABLE users ADD COLUMN form_submitted TINYINT NOT NULL DEFAULT 0 COMMENT '0: Not submitted, 1: Submitted';

-- +goose Down
ALTER TABLE users DROP COLUMN form_submitted;
