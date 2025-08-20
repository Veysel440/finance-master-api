-- +goose Up
CREATE TABLE users (
                       id BIGINT PRIMARY KEY AUTO_INCREMENT,
                       name VARCHAR(120) NOT NULL,
                       email VARCHAR(191) NOT NULL UNIQUE,
                       pass_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE sessions (
                          id BIGINT PRIMARY KEY AUTO_INCREMENT,
                          user_id BIGINT NOT NULL,
                          refresh_hash VARCHAR(255) NOT NULL,
                          expires_at DATETIME NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          INDEX(user_id),
                          CONSTRAINT fk_sess_user FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE TABLE wallets (
                         id BIGINT PRIMARY KEY AUTO_INCREMENT,
                         user_id BIGINT NOT NULL,
                         name VARCHAR(80) NOT NULL,
                         currency CHAR(3) NOT NULL,
                         CONSTRAINT fk_w_user FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE TABLE categories (
                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                            user_id BIGINT NOT NULL,
                            name VARCHAR(80) NOT NULL,
                            type ENUM('income','expense') NOT NULL,
                            CONSTRAINT fk_c_user FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE TABLE transactions (
                              id BIGINT PRIMARY KEY AUTO_INCREMENT,
                              user_id BIGINT NOT NULL,
                              wallet_id BIGINT NOT NULL,
                              category_id BIGINT NOT NULL,
                              type ENUM('income','expense') NOT NULL,
                              amount DECIMAL(14,2) NOT NULL,
                              currency CHAR(3) NOT NULL,
                              note VARCHAR(255),
                              occurred_at DATETIME NOT NULL,
                              updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                              deleted_at DATETIME NULL,
                              INDEX idx_tx_user_updated (user_id, updated_at),
                              CONSTRAINT fk_t_user FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE TABLE audit_logs (
                            id BIGINT PRIMARY KEY AUTO_INCREMENT,
                            user_id BIGINT,
                            action VARCHAR(40),
                            entity VARCHAR(40),
                            entity_id BIGINT,
                            diff_json JSON,
                            at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE audit_logs, transactions, categories, wallets, sessions, users;