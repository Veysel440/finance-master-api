-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
                                          id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                          user_id BIGINT NOT NULL,
                                          action VARCHAR(64) NOT NULL,
                                          entity VARCHAR(64) NOT NULL,
                                          entity_id BIGINT NULL,
                                          details JSON NULL,
                                          created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                          PRIMARY KEY (id),
                                          INDEX idx_audit_user_time (user_id, created_at),
                                          INDEX idx_audit_entity (entity, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +goose Down
DROP TABLE IF EXISTS audit_logs;