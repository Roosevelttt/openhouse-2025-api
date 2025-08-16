-- +goose Up
CREATE TABLE IF NOT EXISTS admins (
  id CHAR(36) NOT NULL PRIMARY KEY,
  ukm_id CHAR(36) NULL,
  division_id CHAR(36) NULL,
  name VARCHAR(255) NOT NULL,
  nrp VARCHAR(9) NOT NULL UNIQUE,
  field VARCHAR(255) NOT NULL,
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  CONSTRAINT fk_admins_ukm_id FOREIGN KEY (ukm_id) REFERENCES ukms(id) ON DELETE CASCADE,
  CONSTRAINT fk_admins_division_id FOREIGN KEY (division_id) REFERENCES divisions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS admins;

