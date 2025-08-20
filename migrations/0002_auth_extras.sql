CREATE TABLE IF NOT EXISTS sessions (
                                        id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                        user_id BIGINT NOT NULL,
                                        refresh_hash VARCHAR(255) NOT NULL,
                                        expires_at DATETIME NOT NULL,
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        INDEX(user_id),
                                        CONSTRAINT fk_sess_user FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_devices (
                                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                            user_id BIGINT NOT NULL,
                                            device_id VARCHAR(191) NOT NULL,
                                            name VARCHAR(120) NOT NULL,
                                            last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                            UNIQUE KEY uniq_user_device (user_id, device_id),
                                            CONSTRAINT fk_dev_user FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS totp_secrets (
                                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                                            user_id BIGINT NOT NULL,
                                            secret VARCHAR(64) NOT NULL,
                                            confirmed_at DATETIME NULL,
                                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                            INDEX(user_id),
                                            CONSTRAINT fk_totp_user FOREIGN KEY(user_id) REFERENCES users(id)
);
-- +goose Down
DROP TABLE totp_secrets, user_devices, sessions;