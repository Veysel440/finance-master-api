-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
                                        id           BIGINT AUTO_INCREMENT PRIMARY KEY,
                                        user_id      BIGINT      NOT NULL,
                                        refresh_hash CHAR(64)    NOT NULL UNIQUE,
                                        ua           VARCHAR(255) NOT NULL,
                                        ip           VARCHAR(64)  NOT NULL,
                                        created_at   DATETIME     NOT NULL,
                                        last_used_at DATETIME     NOT NULL,
                                        expires_at   DATETIME     NOT NULL,
                                        INDEX idx_sessions_user (user_id),
                                        CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id)
                                            ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS sessions;