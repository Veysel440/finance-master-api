-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
                                          id         BIGINT AUTO_INCREMENT PRIMARY KEY,
                                          user_id    BIGINT NULL,
                                          action     VARCHAR(64)  NOT NULL,
                                          entity     VARCHAR(64)  NOT NULL,
                                          entity_id  BIGINT NULL,
                                          details    JSON NULL,
                                          created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                          INDEX idx_audit_user_created (user_id, created_at),
                                          CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users(id)
                                              ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Transactions FK'leri: cüzdan ve kategori silinince SİLMEYİ ENGELLE (RESTRICT).
ALTER TABLE transactions
    ADD CONSTRAINT fk_tx_wallet
        FOREIGN KEY (wallet_id) REFERENCES wallets(id)
            ON DELETE RESTRICT ON UPDATE CASCADE,
    ADD CONSTRAINT fk_tx_category
        FOREIGN KEY (category_id) REFERENCES categories(id)
            ON DELETE RESTRICT ON UPDATE CASCADE;

-- Listeleme ve özet için kombine indeks
CREATE INDEX idx_tx_user_date_type ON transactions(user_id, occurred_at, type);

-- Not alanında fulltext (isteğe bağlı)
CREATE FULLTEXT INDEX idx_tx_note_ft ON transactions(note);

-- Wallets / Categories yardımcı indeksler
CREATE INDEX idx_wallets_user ON wallets(user_id);
CREATE INDEX idx_categories_user_type ON categories(user_id, type);

-- +goose Down
ALTER TABLE transactions
    DROP FOREIGN KEY fk_tx_wallet,
    DROP FOREIGN KEY fk_tx_category;

DROP INDEX idx_tx_user_date_type ON transactions;
DROP INDEX idx_tx_note_ft ON transactions;
DROP INDEX idx_wallets_user ON wallets;
DROP INDEX idx_categories_user_type ON categories;

DROP TABLE IF EXISTS audit_logs;