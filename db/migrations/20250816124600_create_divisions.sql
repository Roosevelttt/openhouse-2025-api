-- +goose Up
CREATE TABLE IF NOT EXISTS divisions (
  id CHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  KEY idx_divisions_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS divisions;

