-- +goose Up
CREATE TABLE IF NOT EXISTS detail_registrations (
  id CHAR(36) NOT NULL PRIMARY KEY,
  nrp VARCHAR(9) NOT NULL,
  ukm_id CHAR(36) NOT NULL,
  payment VARCHAR(255) NULL,
  drive_url VARCHAR(255) NULL,
  file_validated TINYINT NOT NULL COMMENT '0: No, 1: Yes, 2: Reject',
  payment_validated TINYINT NOT NULL COMMENT '0: No, 1: Yes',
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  CONSTRAINT fk_detail_registrations_nrp FOREIGN KEY (nrp) REFERENCES users(nrp) ON DELETE CASCADE,
  CONSTRAINT fk_detail_registrations_ukm_id FOREIGN KEY (ukm_id) REFERENCES ukms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS detail_registrations;

