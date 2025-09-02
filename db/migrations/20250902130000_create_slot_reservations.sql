-- +goose Up
-- Create slot_reservations table for temporary slot holding
CREATE TABLE slot_reservations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    reservation_id VARCHAR(100) UNIQUE NOT NULL,
    nrp CHAR(10) NOT NULL,
    ukm_id CHAR(36) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_reservations_nrp (nrp),
    INDEX idx_reservations_ukm_id (ukm_id),
    INDEX idx_reservations_expires (expires_at),
    INDEX idx_reservations_composite (nrp, ukm_id, expires_at),
    
    FOREIGN KEY (nrp) REFERENCES users(nrp) ON DELETE CASCADE,
    FOREIGN KEY (ukm_id) REFERENCES ukms(id) ON DELETE CASCADE
);

-- Add cleanup event for expired reservations
SET GLOBAL event_scheduler = ON;

CREATE EVENT IF NOT EXISTS cleanup_expired_reservations
ON SCHEDULE EVERY 1 MINUTE
DO
  DELETE FROM slot_reservations WHERE expires_at < NOW();

-- +goose Down
DROP EVENT IF EXISTS cleanup_expired_reservations;
DROP TABLE IF EXISTS slot_reservations;
