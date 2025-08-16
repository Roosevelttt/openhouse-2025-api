-- +goose Up
CREATE TABLE IF NOT EXISTS ukms (
  id CHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  current_slot INT NOT NULL,
  max_slot INT NOT NULL,
  regist_fee INT NOT NULL,
  description TEXT NULL,
  logo_url VARCHAR(255) NULL,
  poster_url VARCHAR(255) NULL,
  UNIQUE KEY ukms_slug_unique (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS ukms;

